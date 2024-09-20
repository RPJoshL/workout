#!/bin/bash

# Script that changes the android ID from ".android." to ".testandroid." to debug the apps
# without uninstalling the main ones

if [ "$1" = "-h" ] || [ "$1" = "--help" ]; then
	echo "default -> test    = androidId.sh test"
	echo "test    -> default = androidId.sh"
	exit 0
fi

if [ "$1" = "test" ]; then
echo "Changing from default -> test"
	grep -rl 'de.rpjosh.rpout.android' ./android/ | xargs --no-run-if-empty sed -i 's/de.rpjosh.rpout.android/de.rpjosh.rpout.testandroid/g'
	grep -rl '/src/main/java/de/rpjosh/rpout/android' ./android/ | xargs --no-run-if-empty sed -i 's/\/src\/main\/java\/de\/rpjosh\/rpout\/android/\/src\/main\/java\/de\/rpout\/rpdb\/testandroid/g'
	grep -rl '<string name="app_name">RPout</string>' ./android/ | xargs --no-run-if-empty sed -i 's|<string name="app_name">RPout</string>|<string name="app_name">RPout (test)</string>|g'
	
	# Android
	mv ./android/app/src/main/java/de/rpjosh/rpout/android/ ./android/app/src/main/java/de/rpjosh/rpout/testandroid/

	# Wear OS
	mv ./android/wear/src/main/java/de/rpjosh/rpout/android/ ./android/wear/src/main/java/de/rpjosh/rpout/testandroid/

	# Shared
	mv ./android/shared/src/main/java/de/rpjosh/rpout/android/ ./android/shared/src/main/java/de/rpjosh/rpout/testandroid/
	mv ./android/shared/schemas/de.rpjosh.rpout.android.shared.persistence.Database/ ./android/shared/schemas/de.rpjosh.rpout.testandroid.shared.persistence.Database/ 

	# Kill any running gradle daemons
	pkill -f '/wrapper/dists/gradle-'
	rm -rf ./android/.gradle/8.7/executionHistory/executionHistory.bin

	# Remove build cache
	rm -rf ./android/shared/build ./android/wear/build ./android/app/build
elif [ "$1" = "-h" ] || [ "$1" = "--help" ] || [ "$1" = "?" ]; then
	echo "Provide 'test' as an argument to change from 'default -> test'"
else 
echo "Changing from test -> default"
	grep -rl de.rpjosh.rpout.testandroid ./android/ | xargs --no-run-if-empty sed -i 's/de.rpjosh.rpout.testandroid/de.rpjosh.rpout.android/g'
	grep -rl '/src/main/java/de/rpjosh/rpout/testandroid' ./android/ | xargs --no-run-if-empty sed -i 's/\/src\/main\/java\/de\/rpjosh\/rpout\/testandroid/\/src\/main\/java\/de\/rpout\/rpdb\/android/g'
	grep -rl '<string name="app_name">RPout (test)</string>' ./android/ | xargs --no-run-if-empty sed -i 's|<string name="app_name">RPout (test)</string>|<string name="app_name">RPout</string>|g'
	
	# Android
	mv ./android/app/src/main/java/de/rpjosh/rpout/testandroid/ ./android/app/src/main/java/de/rpjosh/rpout/android/

	# Wear OS
	mv ./android/wear/src/main/java/de/rpjosh/rpout/testandroid/ ./android/wear/src/main/java/de/rpjosh/rpout/android/

	# Shared
	mv ./android/shared/src/main/java/de/rpjosh/rpout/testandroid/ ./android/shared/src/main/java/de/rpjosh/rpout/android/
	mv ./android/shared/schemas/de.rpjosh.rpout.testandroid.shared.persistence.Database/ ./android/shared/schemas/de.rpjosh.rpout.android.shared.persistence.Database/ 

	# Kill any running gradle daemons
	pkill -f '/wrapper/dists/gradle-'
	rm -rf ./android/.gradle/8.7/executionHistory/executionHistory.bin

	# Remove build cache
	rm -rf ./android/shared/build ./android/wear/build ./android/app/build
fi