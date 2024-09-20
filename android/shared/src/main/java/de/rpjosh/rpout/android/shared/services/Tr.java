package de.rpjosh.rpout.android.shared.services;

import java.util.ArrayList;

/**
 * A static class for the main translations
 */
public class Tr {

    public final static TranslationService translationService = new TranslationService("translation.shared");
    private final static ArrayList<TranslationService> translationServices = new ArrayList<TranslationService>();
    public static TranslationService.Language forcedLanguage = null;

    /**
     * {@link de.rpjosh.rpout.android.shared.services.TranslationService#get(String, boolean, String...)}
     */
    public static String get(String property, boolean capitalize, String... replaceStrings) {
        if (translationService.isPropAvailable(property))
            return translationService.get(property, capitalize, replaceStrings);

        TranslationService tr = translationServices.stream().filter(tr2 -> tr2.isPropAvailable(property)).findFirst().orElse(null);
        if (tr != null) return tr.get(property, capitalize, replaceStrings);
        else			return translationService.get(property, capitalize, replaceStrings);	// this will throw an exception internal
    }

    /**
     * {@link de.rpjosh.rpout.android.shared.services.TranslationService#get(String, boolean, String...)}
     */
    public static String get(String property, String... replaceStrings) {
        return get(property, false, replaceStrings);
    }

    /**
     * Forces the use of the given language for all registered translation services
     *
     * @param language	language to force
     */
    public static void setLanguage(TranslationService.Language language) {
        translationService.setLanguage(language);
        translationServices.forEach(tr -> tr.setLanguage(language));
        forcedLanguage = language;
    }

    public static TranslationService.Language getUsedLanguage() {
        return translationService.getLanguage();
    }

    public static void addTranslationService(TranslationService tr) {
        translationServices.add(tr);
        if (forcedLanguage != null) tr.setLanguage(forcedLanguage);
    }

}
