package de.rpjosh.rpout.android.shared.inject;

import java.lang.reflect.Constructor;
import java.lang.reflect.Field;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.HashMap;
import java.util.List;

import de.rpjosh.rpout.android.shared.services.Logger;

public class InjectionFactory {

    @Inject(parameters= {"InjectionFactory"})
    private Logger logger;

    HashMap<Class<?>, Object> concreteObjects = new HashMap<Class<?>, Object>();
    HashMap<Class<?>, Class<?>> concreteClasses = new HashMap<Class<?>, Class<?>>();
    HashMap<Class<?>, Class<?>> concreteClassesR = new HashMap<Class<?>, Class<?>>();


    public InjectionFactory(Logger logger) {
        this.logger = logger;
    }

    /**
     * Adds a concrete Object for injection for the given class or interface.
     * Any injections present in this object will be ignored
     *
     * @param classT		class or interface for which the object should be injected
     * @param object		object to inject
     */
    public void addConcreteDependency (Class<?> classT, Object object) {
        concreteObjects.put(classT, object);
    }

    /**
     * Adds a concrete Class for injection for the given interface or abstract class
     *
     * @param classInterface	interface or abstract class for which the object should be injected
     * @param concreteClass		class from which an object should be created
     */
    public void addConcreteClass (Class<?> classInterface, Class<?> concreteClass) {
        concreteClasses.put(classInterface, concreteClass);
        concreteClassesR.put(concreteClass, classInterface);
    }

    public void setLogger(Logger logger) {
        this.logger = logger;
    }

    /**
     * Performs an recursive injection for each annotated field in the classes
     *
     * @param objClass					class to create and object and inject the properties
     * @param parameters				parameters for instantiating
     * @param addConcreteDependency		add this object to the concrete dependency configuration
     * @param obj						instead of creating a new instance of the given class,
     * 									the given object will be used to inject the properties. This can be null
     *
     * @return							the instantiated "main" object with the applied dependencies
     */
    @SuppressWarnings("unchecked")
    public <T> T inject(Class<T> objClass, Object[] parameters, boolean addConcreteDependency, T obj) {
        try {
            if (objClass == null && obj != null) {
                objClass = (Class<T>) obj.getClass();
            }

            T concreteObject = getConcreteDependency(objClass);
            if (concreteObject != null) {
                return concreteObject;
            }

            // create a new object
            T object;
            if (obj == null) {
                if (parameters == null || parameters.length == 0) {
                    Constructor<T> constructor = objClass.getDeclaredConstructor();
                    constructor.setAccessible(true);	// also try to call private constructors
                    object = constructor.newInstance();
                } else {
                    Constructor<T> constructor  = objClass.getDeclaredConstructor(Arrays.stream(parameters).map(p -> p.getClass()).toArray(Class[]::new));
                    object = constructor.newInstance((Object[]) parameters);
                }
            } else {
                object = obj;
            }

            for (Field field: getFields(objClass)) {
                if (! field.isAnnotationPresent(Inject.class)) continue;

                Inject inject = field.getAnnotation(Inject.class);

                try {
                    field.setAccessible(true);

                    Object concreteObjectField = getConcreteDependency(field.getType());
                    if (concreteObjectField != null) {
                        field.set(object, concreteObjectField);
                    } else {
                        if (addConcreteDependency)	concreteObjects.put(objClass, object);
                        // set the value of the field recursively
                        if (concreteClasses.containsKey(field.getType()))
                            field.set(object, this.inject(concreteClasses.get(field.getType()), inject.parameters(), false));
                        else
                            field.set(object, this.inject(field.getType(), inject.parameters(), false));
                    }

                } catch (Exception ex) {
                    logger.log("e", ex);
                }
            }

            // call after Inject method
            if (object instanceof Injectable) ((Injectable) object).afterInject();

            if (addConcreteDependency && !concreteObjects.containsKey(objClass))
                concreteObjects.put(objClass, object);

            return object;
        } catch (Exception ex) {

            // it's possible that the logger isn't yet available -> validate
            if (logger == null)					ex.printStackTrace();
            else if (!logger.isLoggerNull())	logger.log("e", ex);
            else								ex.printStackTrace();
            return null;
        }

    }


    @SuppressWarnings("unchecked")
    private <T> T getConcreteDependency(Class<T> cl) {
        // only two level resolve
        if (concreteObjects.containsKey(cl)) {
            return (T) concreteObjects.get(cl);
        }
        if (concreteClasses.containsKey(cl) && concreteObjects.containsKey(concreteClasses.get(cl))) {
            return (T) concreteObjects.get(concreteClasses.get(cl));
        }
        if (concreteClassesR.containsKey(cl) && concreteObjects.containsKey(concreteClassesR.get(cl))) {
            return (T) concreteObjects.get(concreteClassesR.get(cl));
        }
        return null;
    }

    /**
     * {@link #inject(Class, String[], boolean)}
     */
    public <T> T inject(Class<T> objClass, Object[] parameters, boolean addConcreteDependency) {
        return inject(objClass, parameters, addConcreteDependency, null);
    }

    /**
     * {@link #inject(Class, String[], boolean, Object)}
     */
    public <T> T inject(Class<T> objClass, String[] parameters, boolean addConcreteDependency) {
        return inject(objClass, parameters, addConcreteDependency, null);
    }

    /**
     * Gets all fields from the given class and the super classes
     *
     * @param objClass	the class to extract the fields
     * @return			all fields
     */
    private List<Field> getFields(Class<?> objClass) {
        List<Field> fields = new ArrayList<>();
        while (objClass != Object.class) {
            fields.addAll(Arrays.asList(objClass.getDeclaredFields()));
            objClass = objClass.getSuperclass();
        }
        return fields;
    }



}
