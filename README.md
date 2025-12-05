## Workout

A simple workout tracking web application for GPX based activities.

## Create user

A new user can only be created over the CLI.

```
./workout user create
```

It's important that the new user updates it's profile in the settings so activity indicators and calories are calculated correctly.
Otherwise, you can also use the flag `allFields`.

### Known issues

* Leaflet tooltip stuck while panning: [Is Fixed in main](https://github.com/Leaflet/Leaflet/pull/9154)

### To-Do

* Complete the light mode
* Use [`Leaflet.markercluster`](https://github.com/Leaflet/Leaflet.markercluster) in workout search for high zoom levels and many workouts