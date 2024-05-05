## Workout

A simple workout tracking web application for GPX based activities.


### User data

To calculate 

#### Personal body data 

To calculate some activity indicators and your calories burned during workout, you have to provide the following data:

* Weight in kg
* Height in cm
* Birth year for calculating your age
* VO2max value. It is the maximum amount of oxygen that your body can effectively use during one minute of physical activity *(mL/kg/min)*. If you don't have a smart watch that can calculate this value, you can do the `Coopers-Test`


Entfernung zwischen zwei PUnkten: https://gist.github.com/hotdang-ca/6c1ee75c48e515aec5bc6db6e3265e49
Städte im Umkreis: https://download.geonames.org/export/dump/cities1000.zip mit Geonames parser: https://github.com/mkrou/geonames
Workouts im Umkreis: https://stackoverflow.com/questions/42799118/mysql-find-points-within-radius-from-database
Kalorien berechnen: https://www.omnicalculator.com/sports/calories-burned-by-heart-rate
Hier abkupfern: https://github.com/jovandeginste/workout-tracker

Punkte Downsampel mit: Ramer–Douglas–Peucker

### Known issues

* Leaflet tooltip stuck while panning: [Is Fixed in main](https://github.com/Leaflet/Leaflet/pull/9154)

### To-Do

* Use [`Leaflet.markercluster`](https://github.com/Leaflet/Leaflet.markercluster) in workout search for high zoom levels and many workouts 