# even if there are warnings, the .jar file will be created
-ignorewarnings

# Keep some things that are needed for the app to work correctly
-keepattributes Exceptions,*Annotation*,Signature,!LocalVariableTable,!LocalVariableTypeTable

# Fore a better debugging experience, keep the lines and source file
-keepattributes SourceFile,LineNumberTable
# Rename all source files to "SourceFile"
#-renamesourcefileattribute SourceFile

# Keep constructor
-keepclassmembers public class * {
    public <init>(...);
}

# optimize code
#-dontobfuscate
-optimizations !code/simplification/arithmetic,!code/simplification/cast,!field/*,!class/merging/*
-optimizationpasses 5
-allowaccessmodification
-dontskipnonpubliclibraryclasses
-repackageclasses 'o'

-keep class javax.** { *; }
-keep class java.** { *; }
-keep class androidx.health.services.client.proto.** { *; }

-keepclasseswithmembernames class * {
    native <methods>;
}
-keepclasseswithmembers public class * {
	public static void main(java.lang.String[]);
}

-keepclassmembers enum * {
    public static **[] values();
    public static ** valueOf(java.lang.String);
}

-keep class  de.rpjosh.rpout.android.shared.models** { *; }

-keep interface de.rpjosh.rpout.android.shared.inject.Inject
-keepclassmembers class * {
	@de.rpjosh.rpout.android.shared.inject.Inject *;
}

-keep class ch.qos.logback.** { *; }
-keep class retrofit2.** { *; }
-keep class com.sun.** { *; }

# Keep GSON stuff
-keep class com.google.gson.reflect.TypeToken
-keep class * extends com.google.gson.reflect.TypeToken
-keep class * implements java.lang.reflect.Type

-dontwarn javax.annotation.**
-dontwarn kotlinx.**
-dontwarn kotlin.**


-keepclassmembers class * implements java.io.Serializable {
	static final long serialVersionUID;
	static final java.io.ObjectStreamField[] serialPersistentFields;
	private void writeObject(java.io.ObjectOutputStream);
	private void readObject(java.io.ObjectInputStream);
	java.lang.Object writeReplace();
	java.lang.Object readResolve();
}

#-dontusemixedcaseclassnames
#-dontskipnonpubliclibraryclasses
#-verbose

## Androids default proguard file ##
-keepattributes *Annotation*

# For native methods, see http://proguard.sourceforge.net/manual/examples.html#native
-keepclasseswithmembernames class * {
    native <methods>;
}

# keep setters in Views so that animations can still work.
# see http://proguard.sourceforge.net/manual/examples.html#beans
-keepclassmembers public class * extends android.view.View {
   void set*(***);
   *** get*();
}

# We want to keep methods in Activity that could be used in the XML attribute onClick
-keepclassmembers class * extends android.app.Activity {
   public void *(android.view.View);
}

# For enumeration classes, see http://proguard.sourceforge.net/manual/examples.html#enumerations
-keepclassmembers enum * {
    public static **[] values();
    public static ** valueOf(java.lang.String);
}

-keepclassmembers class * implements android.os.Parcelable {
  public static final android.os.Parcelable$Creator CREATOR;
}

-keepclassmembers class **.R$* {
    public static <fields>;
}

# The support library contains references to newer platform versions.
# Don't warn about those in case this app is linking against an older
# platform version.  We know about them, and they are safe.
-dontwarn android.support.**