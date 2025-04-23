## Show notifications in workout tracking screen

You have to send an intent like the example below to create a new notification.

```sh
adb shell am broadcast \
    -a de.rpjosh.rpout.android.workout.NOTIFICATION \
    --es notification_action "CREATE" \
	--el notification_id "2" \
	--es notification_category "someMetadata" \
    -p de.rpjosh.rpout.android
```

To delete this notification again, you can send an intent with the action `DELETE`.

```sh
adb shell am broadcast \
    -a de.rpjosh.rpout.android.workout.NOTIFICATION \
    --es notification_action "DELETE" \
	--el notification_id "2" \
    -p de.rpjosh.rpout.android
```