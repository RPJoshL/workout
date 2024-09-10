## Workout

A simple workout tracking web application for GPX based activities.


### User data

To calculate 

#### Personal body data 

To calculate some activity indicators and your calories burned during workout, you have to provide the following data:

* Weight in kg
* Height in cm
* Birth year for calculating your age
* VO2max value. It is the maximum amount of oxygen that your body can effectively use during one minute of physical activity *(mL/kg/min)*. If you don't have a smartwatch that can calculate this value, you can do the `Coopers-Test`

Wie erstellen?

```
Code für die Kommandozeile
```

### Known issues

* Leaflet tooltip stuck while panning: [Is Fixed in main](https://github.com/Leaflet/Leaflet/pull/9154)

### To-Do

* Add workout statistics
	* Outsource search from overview into a generic component
	* Show various activity indexes and metrics
	* Show steps and PAI score
* Complete the light mode
* Document uploading of raw files *(extend API documentation)* and adjust uploader auth mechanism
* Use [`Leaflet.markercluster`](https://github.com/Leaflet/Leaflet.markercluster) in workout search for high zoom levels and many workouts