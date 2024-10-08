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

    private final GlobalConfiguration globalConfig;

    /**
     * Creates a new filter for stderr and stdout
     *
     * @param   globalConfig	the global configuration of the program to prevent printing to the console on quiet
     */
    public LogFilter(GlobalConfiguration globalConfig) {
        this.globalConfig = globalConfig;
    }

    @Override
    public FilterReply decide(ILoggingEvent iLoggingEvent) {
        if (globalConfig.getLogLevel().value * 1000 > iLoggingEvent.getLevel().levelInt) return FilterReply.DENY;
        else return FilterReply.ACCEPT;
    }
}
