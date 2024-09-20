package de.rpjosh.rpout.android

import de.rpjosh.rpout.android.shared.services.MessageType

interface WearMessageReceiver {
    fun onWearMessageReceived(type: MessageType, data: String)
}