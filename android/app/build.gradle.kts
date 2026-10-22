import java.util.Properties

plugins {
    id("com.android.application")
    id("org.jetbrains.kotlin.android")
}

// Release signing comes from android/keystore.properties (never committed):
// storeFile, storePassword, keyAlias, keyPassword. Without it, release builds use
// the debug key, which cannot be upgraded by a later release-signed build.
val keystoreProps = rootProject.file("keystore.properties").takeIf { it.isFile }?.let { file ->
    Properties().apply { file.inputStream().use(::load) }
}

android {
    namespace = "io.github.kevsosmooth.tvremote"
    compileSdk = 36

    defaultConfig {
        applicationId = "io.github.kevsosmooth.tvremote"
        minSdk = 26
        targetSdk = 36
        versionCode = 1
        versionName = "0.1.0"
    }

    signingConfigs {
        if (keystoreProps != null) {
            create("release") {
                storeFile = rootProject.file(keystoreProps.getProperty("storeFile"))
                storePassword = keystoreProps.getProperty("storePassword")
                keyAlias = keystoreProps.getProperty("keyAlias")
                keyPassword = keystoreProps.getProperty("keyPassword")
            }
        }
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"))
            signingConfig = signingConfigs.getByName(if (keystoreProps != null) "release" else "debug")
        }
    }

    packaging {
        // tvlink ships as libtvlink.so and is run as a program, so it must be
        // extracted to nativeLibraryDir (the only app path Android lets us exec).
        jniLibs.useLegacyPackaging = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}
