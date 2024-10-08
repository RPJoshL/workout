package de.rpjosh.rpout.android.shared.services;

import android.util.Log;

import java.io.File;
import java.io.FileWriter;
import java.io.IOException;
import java.io.PrintWriter;
import java.io.RandomAccessFile;
import java.io.StringWriter;
import java.nio.file.Files;
import java.text.MessageFormat;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

import org.slf4j.LoggerFactory;

import ch.qos.logback.classic.ClassicConstants;
import ch.qos.logback.classic.Level;
import ch.qos.logback.classic.LoggerContext;
import ch.qos.logback.classic.encoder.PatternLayoutEncoder;
import ch.qos.logback.classic.spi.ILoggingEvent;
import ch.qos.logback.core.ConsoleAppender;
import ch.qos.logback.core.rolling.FixedWindowRollingPolicy;
import ch.qos.logback.core.rolling.RollingFileAppender;
import ch.qos.logback.core.rolling.SizeBasedTriggeringPolicy;
import ch.qos.logback.core.util.FileSize;
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration;
import de.rpjosh.rpout.android.shared.inject.Inject;
import de.rpjosh.rpout.android.shared.inject.Injectable;

public class Logger implements Injectable {


    public enum LEVEL {

        DEBUG(10, "DEBUG", "debug"), INFO(20, "INFO ", "info"), WARNING(30, "WARN ", "warning"),
        ERROR(40, "ERROR", "error"), ERROR_PRINT(40, "ERROR", "errorPrint");

        public final int value;
        public final String output;
        public final String label;

        LEVEL(int value, String output, String label) {
            this.value = value;
            this.output = output;
            this.label = label;
        }
    }

    @Inject GlobalConfiguration config;
    @Inject private ResponseViewInterface response;

    private final String classLocation;
    private ch.qos.logback.classic.Logger logger;


    public Logger(String classLocation) {
        this.classLocation = classLocation;
    }
    @Override
    public void afterInject() {
        createLogger(false);
    }

    // the file appender can only be created once -> make it static
    private static RollingFileAppender<ILoggingEvent> fileAppender;
    private static final Object fileAppenderLock = new Object();

    private void createLogger(boolean initialTry) {

        LoggerContext context = (LoggerContext) LoggerFactory.getILoggerFactory();
        PatternLayoutEncoder layout = new PatternLayoutEncoder();

        layout.setContext(context);
        String normalPattern = "%-12date{YYYY-MM-dd HH:mm:ss.SSS} [%-5level] %-15.15([%thread)] %-25.25([%replace(%replace(%file){'.java', ''}){'.kt', ''}:%line)] - %msg%n";

        // when the code got obfuscated, the source name won't contain logger anymore -> use the provided name
        if (!this.getClass().getCanonicalName().contains("ogger")) {
            normalPattern = normalPattern.replace("(%file)", "(%logger)");
        }

        layout.setPattern(normalPattern);
        layout.start();
        System.setProperty(ClassicConstants.LOGBACK_CONTEXT_SELECTOR, Logger.class.getName());
        synchronized(fileAppenderLock) {
            if (fileAppender == null) {
                PatternLayoutEncoder layoutFile = new PatternLayoutEncoder();
                layoutFile.setContext(context);
                layoutFile.setPattern(normalPattern);
                layoutFile.start();

                fileAppender = new RollingFileAppender<ILoggingEvent>();
                fileAppender.setContext(context);
                fileAppender.setName("RPout.log");
                fileAppender.setEncoder(layoutFile);
                fileAppender.setAppend(true);

                // logging file can at a maximum be 750kb big | 3 backups are kept
                // when the config has a dependency to the logger, the configuration directory cannot be determined - unfortunately
                try { if (config.getAppDir("logs") == null)	throw new Exception(); }
                catch (Exception ex) {
                    fileAppender = null; 	// we need to set it to null again that the file appender will created again
                    if (initialTry) { printToConsole(LEVEL.ERROR, "Cannot create logger...\n" + getStackTrace(ex), true); }
                    return;
                }
                fileAppender.setFile(config.getAppDir("logs") + config.getApplicationName() + ".log");
                FixedWindowRollingPolicy rollover = new FixedWindowRollingPolicy();
                rollover.setContext(context); rollover.setParent(fileAppender);
                rollover.setFileNamePattern(config.getAppDir("logs") + config.getApplicationName() + "-%i.log");
                rollover.setMinIndex(1);
                rollover.setMaxIndex(3);
                rollover.start();
                fileAppender.setRollingPolicy(rollover);

                SizeBasedTriggeringPolicy<ILoggingEvent> trigger = new SizeBasedTriggeringPolicy<>();
                trigger.setContext(context);
                trigger.setMaxFileSize(new FileSize(750000));
                trigger.start();
                fileAppender.setTriggeringPolicy(trigger);
                fileAppender.addFilter(new LogFilter(config));

                fileAppender.start();
            }
        }

        logger = context.getLogger(classLocation);
        logger.setLevel(Level.DEBUG);
        logger.setAdditive(false);	// Print to root logger?
        logger.addAppender(fileAppender);
    }

    /**
     * Get log file returns a file that contains the last 1000 lines of logs
     *
     * @return      File with the log entries. This temporary file should be deleted
     *
     * @throws Exception
     */
    public File getLogFile() throws  Exception {
        return getLogFile(1000);
    }

    /**
     * Get log file returns a file that contains the last x lines of logs
     *
     * @param maxLines      Maximum lines to return
     *
     * @return              File with the log entries. This temporary file should be deleted
     *
     * @throws Exception
     */
    public File getLogFile(int maxLines) throws Exception {
        // Try to read from last two log files
        File file1 = new File(config.getAppDir("logs") + "RPout.log");
        File file2 = new File(config.getAppDir("logs") + "RPout-1.log");

        // Create a temp file to write the combined log to
        File tempFile = File.createTempFile("RPout_tmp", ".txt");

        /* Number of written lines */
        int i = 0;
        try (FileWriter fw = new FileWriter(tempFile)) {
            // Read complete first file
            List<String> firstFileLines = new ArrayList<>();
            if (file1.exists() && file1.isFile()) {
                String firstFile = new String(Files.readAllBytes(file1.toPath()));
                firstFileLines = new ArrayList<>(Arrays.asList(firstFile.split("\n")));

                // Limit max number of files
                i = firstFileLines.size();
                if (i > maxLines + 10) {
                    firstFileLines = firstFileLines.subList(i - maxLines, i);
                }
            }

            // If we don't have 1000 lines already, read text from previous log file
            StringBuilder bd = new StringBuilder();
            if (file2.exists() && file2.isFile()) {
                try (RandomAccessFile raf = new RandomAccessFile(file2, "r")) {
                    for (long pointer = file2.length(); pointer >= 0; pointer--) {
                        raf.seek(pointer);
                        char c = (char) raf.read();

                        // Break when max number of liens were read
                        if (i > maxLines) {
                            break;
                        }

                        // Break for line break
                        if (c == '\n') {
                            i++;
                        }

                        bd.append(c);
                    }
                } catch (Exception ex) {
                    log("e", ex, "Failed to read from log file (2)");
                }

                // Combine lines
                if (bd.length() > 10) {
                    ArrayList<String> firstFileLines2 = new ArrayList<>(Arrays.asList(bd.reverse().toString().split("\n")));
                    firstFileLines2.addAll(firstFileLines);
                    firstFileLines = firstFileLines2;
                }
            }

            // Write all lines to file
            firstFileLines.forEach(line -> {
                try {
                    fw.write(line + "\n");
                } catch (IOException e) {
                    throw new RuntimeException(e);
                }
            });
        } catch (Exception ex) {
            log("e", ex, "Failed to write to temp file");
        }

        return tempFile;
    }

    public boolean isLoggerNull() {
        return logger == null;
    }

    public void log(String level, String message) {
        log(getLevel(level), null, message);
    }
    public void log(String level, String message, Object ...formatingOptions) {
        log(getLevel(level), null, message, formatingOptions);
    }
    public void log(String level, Exception ex) {
        log(getLevel(level), ex, "");
    }
    public void log(String level, Exception ex, String message) {
        log(getLevel(level), ex, message);
    }
    public void log(String level, Exception ex, String message, Object ...formattingOptions) {
        log(getLevel(level), ex, message, formattingOptions);
    }

    private LEVEL getLevel(String levelName) {
        levelName = levelName.toLowerCase().strip();
        return switch (levelName) {
            case "ee" -> LEVEL.ERROR_PRINT;
            case "e" -> LEVEL.ERROR;
            case "w" -> LEVEL.WARNING;
            case "i" -> LEVEL.INFO;
            default -> LEVEL.DEBUG;
        };
    }

    /**
     * Logs an message
     *
     * @param level				level of the information (DEBUG, INFO, WARN, ERROR, ERROR_PRINT)
     * @param ex				exception to log [@nullable]
     * @param message			message to log
     * @param formatingOptions	the message can contain various placeholder ({0} {1}). See MessageFormat for more informations
     */
    private void log(LEVEL level, Exception ex, String message, Object... formatingOptions) {

        if (this.isLoggerNull()) createLogger(true);

        // This could throw an exception
        try {
            if (formatingOptions != null && formatingOptions.length != 0) message = MessageFormat.format(message.replace("'", "''"), formatingOptions);
        } catch (Exception exF) {
            this.log("w", "Failed to format log message", exF);
        }

        if (!this.isLoggerNull()) {
            logger.log(null, Logger.class.getCanonicalName(), level.value, message, new String[] { level.label }, ex);
        }

        // Also display error in UI
        if (response != null && level == LEVEL.ERROR_PRINT) {
            response.displayError(message);
        }

        printToConsole(level,
                //LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSS")) +
                "[" + level.output + "] - " +
                        message + (message.isEmpty() ? "" : "\n") +
                        getStackTrace(ex),
                false
        );
    }

    /** This function doesn't do anything, but can be overwritten by android (no support for console logging) */
    private void printToConsole(LEVEL level, String message, boolean forcePrint) {
        if (message == null) return;

        String tag = "RPout-Logger";
        switch (level) {
            case DEBUG: { Log.d(tag, message); break; }
            case INFO: { Log.i(tag, message); break; }
            case WARNING: { Log.w(tag, message); break; }
            default: { Log.e(tag, message); break; }
        }
    }

    private String getStackTrace(Exception ex) {
        if (ex == null) return "";

        StringWriter sw = new StringWriter();
        PrintWriter pw = new PrintWriter(sw);

        ex.printStackTrace(pw);
        return sw.toString();
    }

}
