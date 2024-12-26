package de.rpjosh.rpout.android.shared.controller

import android.content.Context
import androidx.room.Room
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import de.rpjosh.rpout.android.shared.inject.InjectionFactory
import de.rpjosh.rpout.android.shared.persistence.Database
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.ResponseViewInterface
import de.rpjosh.rpout.android.shared.services.SystemUtilsInterface
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.services.TranslationService
import de.rpjosh.rpout.android.shared.services.WearSynchronizationInterface
import kotlin.reflect.KClass

open class AppController
    (  context: Context,
       response: KClass<out ResponseViewInterface>,
       systemUtils: KClass<out SystemUtilsInterface>,
       dataSync: KClass<out WearSynchronizationInterface>
    ) {

    companion object {
        lateinit var logger: Logger
            private set

        /**
         * Adds additional properties file for the translation support
         *
         * @param properties	extra properties files (eg. <i> translation.headless </i>)
         */
        fun addAdditionalPropsTranslations(vararg properties: String) {
            properties.forEach {
                Tr.addTranslationService(TranslationService(it))
            }
        }

    }

    // Injection factory
    val injection: InjectionFactory = InjectionFactory(Logger(InjectionFactory::class.java.canonicalName))

    // Configuration
    val globalConfiguration: GlobalConfiguration

    // Other dependencies
    val response: ResponseViewInterface
    lateinit var database: Database
    val systemUtils: SystemUtilsInterface

    init {
        // Add concrete dependencies to inject
        injection.addConcreteClass(ResponseViewInterface::class.java, response.java)
        injection.addConcreteClass(SystemUtilsInterface::class.java, systemUtils.java)
        injection.addConcreteClass(WearSynchronizationInterface::class.java, dataSync.java)

        // Inject configuration we need in almost all classed
        globalConfiguration = injection.inject(GlobalConfiguration::class.java, null, true)
        globalConfiguration.appDir = context.filesDir.absolutePath + "/"

        // Create a new database class (early). We have to initialize this synchronously because we need the data for startup.
        // Because only generic interfaces (without any dependencies to this app) are initialized, we can do it here
        val sync = Object()
        Thread {
            synchronized(sync) {
                database = Room.databaseBuilder(
                    context, Database::class.java, "main-db"
                ).build()
                injection.addConcreteDependency(Database::class.java, database)
                initializeUserSetting()
            }

            // Update unfinished workouts. The app probably crashed or the device was shut down
            database.WorkoutDao().finishAllWorkouts()
        }.start()

        beforeInjection()

        // Add dependencies we need a concrete dependency for
        this.response = injection.inject(response.java, null,  true)

        // Inject global logger
        logger = injection.inject(Logger::class.java, arrayListOf("APP").toArray(), false)
        TranslationService.setLogger(injection.inject(Logger::class.java, arrayListOf("Translation service").toArray(), false));

        // Initialize controller
        injection.inject(dataSync.java, null, true)
        this.systemUtils = injection.inject(systemUtils.java, null, true)
        injection.inject(MetricController::class.java, null, true)
        injection.inject(WorkoutController::class.java, null, true)

        // Make sure that settings are initialized. All classes that do have a dependency to a db, has to be initialized inside this block
        synchronized(sync) {
            // Start services
            if (globalConfiguration.user != null) startServices()
        }
    }

    /** Before injection is called before all dependencies are injected (except the global config) */
    protected open fun beforeInjection() {

    }

    /** Loads all user settings from the internal database and stores them in memory */
    protected fun initializeUserSetting() {
        val users = database.userDao().getAll()
        if (users.isNotEmpty()) {
            globalConfiguration.user = users[0]
        }
    }

    /** startServices is called when a user reference exists and all dependencies were injected */
    public fun startServices() {

    }
}