package de.rpjosh.rpout.android.shared.services;

import android.util.Log;

import java.util.Arrays;
import java.util.List;
import java.util.Locale;
import java.util.ResourceBundle;

/**
        * A class providing translation support of properties
 */
public class TranslationService {

    private String resourceFile;

    public enum Language {

        GERMAN (new Locale("de", "DE"), "Deutsch"),
        ENGLISH(new Locale("en", "US"), "English");

        public final Locale locale;
        public final String name;

        Language(Locale locale, String name) {
            this.locale = locale;
            this.name = name;
        }

        public static Language fromAndroidLocale(Locale androidLocale) {
            if (androidLocale == null) {
                return ENGLISH;
            }

            String languageCode = androidLocale.getLanguage();
            for (Language lang : values()) {
                if (lang.locale.getLanguage().equals(languageCode)) {
                    return lang;
                }
            }

            return ENGLISH;
        }

    }

    private static Logger logger;
    private ResourceBundle bundle;
    private ResourceBundle defaultBundle;

    /**
     * Creates a new instance for translating support
     *
     * @param resourceFile		the property file to use for the translations. For example translation.de.rpjosh.installer
     */
    public TranslationService(String resourceFile) {
        this.resourceFile = resourceFile;

        try {
            this.defaultBundle = ResourceBundle.getBundle(resourceFile, Locale.ENGLISH);

            Locale osLocale = Locale.getDefault();
            List<Locale> supportedLanguages = (List<Locale>) Arrays.asList(
                    new Locale[] { Language.GERMAN.locale, Language.ENGLISH.locale }
            );
            if (supportedLanguages.contains(osLocale))  this.bundle = ResourceBundle.getBundle(resourceFile, osLocale);
            else										this.bundle = defaultBundle;
        } catch (Exception ex) {
            ex.printStackTrace();
        }
    }

    /**
     * Creates a new instance for translating and forces the use of a specific language for the translations
     * (defaulting to the language the operation system provides)
     *
     * @param resourceFile	property file to use for the translations. For example translation.de.rpjosh.installer
     * @param language		language to force
     */
    public TranslationService(String resourceFile, Language language) {
        this.resourceFile = resourceFile;

        this.defaultBundle = ResourceBundle.getBundle(resourceFile, Locale.ENGLISH);
        this.bundle = ResourceBundle.getBundle(resourceFile, language.locale);
    }


    /**
     * Force the use of a specific language for the translations (defaulting to
     * the language the operation system provides)
     *
     * @param language		language to force
     */
    public void setLanguage(Language language) {
        if (language == null) {
            Locale osLocale = Locale.getDefault();
            List<Locale> supportedLanguages = (List<Locale>) Arrays.asList(
                    new Locale[] { Language.GERMAN.locale, Language.ENGLISH.locale }
            );
            if (supportedLanguages.contains(osLocale))  this.bundle = ResourceBundle.getBundle(resourceFile, osLocale);
            else										this.bundle = defaultBundle;
        } else {
            this.bundle = ResourceBundle.getBundle(resourceFile, language.locale);
        }
    }

    /**
     * Returns the currently used language
     *
     * @return	currently used language
     */
    public Language getLanguage() {
        if (bundle == null) return Language.ENGLISH;

        String language = bundle.getLocale().getLanguage();
        if (language.equals(Language.ENGLISH.locale.getLanguage())) 	return Language.ENGLISH;
        if (language.equals(Language.GERMAN.locale.getLanguage()))		return Language.GERMAN;

        return Language.ENGLISH;
    }

    /**
     * Returns the translated value of the property
     *
     * @param property		 property to translate
     * @param capitalize     the first letter of the translated string will be capitalized
     * @param replaceStrings strings for replacing {0}, {1}, ... inside the property values starting by zero counting one by one
     *
     * @return 				translated string for the property
     */
    public String get(String property, boolean capitalize, String... replaceStrings) {
        try {
            return replaceValues(property, bundle.getString(property), capitalize, replaceStrings);
        } catch (Exception ex) {
            if (bundle == null && logger == null) Log.d("e", "No bundle and logger. It's probably called from a Compose preview");
            else if (bundle == null) logger.log("w", "Bundle is null");
            else if (logger != null) logger.log("d", "Cannot find property: \"" + property + "\" for language \"" + bundle.getLocale().getLanguage() + "\" in property file \"" + resourceFile + "\"", "Translations#get");
        }

        try {
            return replaceValues(property, defaultBundle.getString(property), capitalize, replaceStrings);
        } catch (Exception ex) {
            if (logger == null) return property;
            logger.log("e", "Cannot find property: \"" + property + "\" in default translation list from property file \"" + resourceFile + "\"", "Translations#get");
            return property;
        }
    }

    private String replaceValues(String property, String value, boolean capitalize, String... replaceStrings) {

        if (capitalize && value.length() > 1) {
            value = value.substring(0,1).toUpperCase() + value.substring(1);
        }
        if (replaceStrings.length == 0) return value;

        for (int i = 0; i < replaceStrings.length; i++) {
            if (value.contains("{" + i + "}")) {
                value = value.replace("{" + i + "}", replaceStrings[i] == null ? "<null>" : replaceStrings[i]);
            } else {
                logger.log("d", "No value matches for " + "{" + i + "}" + "in the property \"" + property + "\n: " + value, "Translations#replaceValue");
            }
        }

        return value;
    }

    /**
     * Checks if the given property is available
     *
     * @param property	property to check
     *
     * @return			if the property is available in this resource file (in the default bundle 'en' or the actual)
     */
    public boolean isPropAvailable(String property) {
        try {
            defaultBundle.getString(property);
            return true;
        } catch (Exception ex) {
            try {
                bundle.getString(property);
                return true;
            } catch (Exception ex2) { }
        }

        return false;
    }

    public static void setLogger(Logger logger) { TranslationService.logger = logger; }

}

