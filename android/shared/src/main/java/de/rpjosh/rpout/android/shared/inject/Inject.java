package de.rpjosh.rpout.android.shared.inject;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Annotates an attribute of the class to be injectable by the InjectionFactory
 */
@Retention(RetentionPolicy.RUNTIME)
@Target({ElementType.FIELD})
public @interface Inject {

    /**
     * Parameters to call the constructor
     */
    public String[] parameters() default {};
}
