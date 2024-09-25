# Build android app
FROM quay.io/alvr/alpine-android:android-34-jdk17-v2023.09.10 AS androidBuilder

WORKDIR /build

# Download and install gradle through gradlew wrapper
ENV ORG_GRADLE_WELCOME=NEVER GRADLE_OPTS="-Xmx96m -Dorg.gradle.configureondemand=true -Dorg.gradle.daemon=false -Dorg.gradle.jvmargs='-Xmx560m -XX:MaxMetaspaceSize=288m' -Dorg.gradle.workers.max=1"
COPY ./android/gradlew .
COPY /android/gradle ./gradle
RUN ./gradlew

ARG VERSION=0.0.0
# Copy all required files
COPY ./VERSION ./VERSION
COPY ./android ./android

# Build (unsigned) APK
RUN export RPDB_CONTAINER_BUILD=true && ./android/gradlew -p ./android/ assembleRelease \
	&& rm -rf ./release/RPout_*-aligned.apk \
	# Copy unaglined APK
	&& cp ./release/RPout_${VERSION}.apk ./release/RPout_${VERSION}-unaligned.apk \
	&& cp ./release/RPout_wear_${VERSION}.apk ./release/RPout_wear_${VERSION}-unaligned.apk \
	# Align APK
	&& zipalign -p 4 ./release/RPout_${VERSION}.apk ./release/RPout_${VERSION}-unsigned.apk > /dev/null \
	&& zipalign -p 4 ./release/RPout_wear_${VERSION}.apk ./release/RPout_wear_${VERSION}-unsigned.apk > /dev/null

# Sign APK
RUN --mount=type=secret,id=androidKeystore --mount=type=secret,id=androidKeystorePassword \
	   java -jar $ANDROID_HOME/build-tools/34.0.0/lib/apksigner.jar sign \
		 --ks /run/secrets/androidKeystore --ks-pass pass:$(cat /run/secrets/androidKeystorePassword) --out ./release/RPout_${VERSION}.apk ./release/RPout_${VERSION}-unsigned.apk \
	&& java -jar $ANDROID_HOME/build-tools/34.0.0/lib/apksigner.jar sign \
	     --ks /run/secrets/androidKeystore --ks-pass pass:$(cat /run/secrets/androidKeystorePassword) --out ./release/RPout_wear_${VERSION}.apk ./release/RPout_wear_${VERSION}-unsigned.apk \
	&& rm -rf ./release/RPout_*.idsig

# For a google play release we need to generate an APP bundle
RUN export RPDB_CONTAINER_BUILD=true && export RPDB_CONTAINER_BUILD_DEP=true \
	&& ./android/gradlew -p ./android/ bundleRelease -PbuildBundle


# Base container to serve the static files from
FROM docker.io/httpd:2.4-bookworm

# Install java (for apksigner)
RUN apt-get update && apt-get install -y openjdk-17-jdk-headless zipalign protobuf-compiler

# Copy apksigner
COPY ./docker/android/apksigner.jar /app/apksigner.jar
COPY ./docker/android/aab/ /app/aab

# Copy downloadable apps
COPY --from=androidBuilder --chmod=0777 /build/release/RPout* /usr/local/apache2/htdocs/

# Copy entrypoint and configuration
COPY --chmod=555 ./docker/android/entrypoint.sh /entrypoint.sh
COPY ./docker/android/httpd.conf /usr/local/apache2/conf/httpd.conf
RUN rm /usr/local/apache2/htdocs/index.html && chmod 0777 /usr/local/apache2/logs && chmod 0777 /usr/local/apache2/htdocs

# Replace version in configuration file
ARG VERSION=0.0.0
RUN sed -i 's/{{VERSION}}/'$VERSION'/g' /usr/local/apache2/conf/httpd.conf
RUN echo $VERSION > /app/VERSION

EXPOSE 8090

# Use some random user
USER 12345

ENTRYPOINT [ "/entrypoint.sh" ]