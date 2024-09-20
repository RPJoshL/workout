package de.rpjosh.rpout.android.shared.services;


import ch.qos.logback.classic.Level;
import ch.qos.logback.classic.spi.ILoggingEvent;
import ch.qos.logback.core.filter.Filter;
import ch.qos.logback.core.spi.FilterReply;
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration;

/**
 * Filters log messages based on stdout or stderr and the configured log level
 */
public class LogFilter extends Filter<ILoggingEvent> {

    private GlobalConfiguration globalConfig;
    private boolean matchStdErr;

    /**
     * Creates a new filter for stderr and stdout
     *
     * @param 	matchStdErr		only levels for stderr (error / warn) will be matched. Otherwise invert the result
     * @param   globalConfig	the global configuration of the program to prevent printing to the console on quiet
     */
    public LogFilter(boolean matchStdErr, GlobalConfiguration globalConfig) {
        this.matchStdErr = matchStdErr;
        this.globalConfig = globalConfig;
    }

    @Override
    public FilterReply decide(ILoggingEvent iLoggingEvent) {

        // Filter the log for output //

        // Don't log "ee" errors because they should be printed out in the GUI, console, ...
        if (iLoggingEvent.getArgumentArray() != null && iLoggingEvent.getArgumentArray()[0].equals("errorPrint")) {
            return FilterReply.DENY;
        }
        if (globalConfig.getLogLevel().value * 1000 > iLoggingEvent.getLevel().levelInt)
            return FilterReply.DENY;

        if (iLoggingEvent.getLevel() == Level.ERROR || iLoggingEvent.getLevel() == Level.WARN) {
            return matchStdErr ? FilterReply.ACCEPT : FilterReply.DENY;
        }

        return matchStdErr ? FilterReply.DENY : FilterReply.ACCEPT;
    }
}
