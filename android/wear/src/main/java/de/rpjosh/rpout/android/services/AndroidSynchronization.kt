package de.rpjosh.rpout.android.services

import android.util.Log
import com.google.android.gms.wearable.CapabilityClient
import com.google.android.gms.wearable.Node
import com.google.android.gms.wearable.Wearable
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.services.WearSynchronizationInterface

/** WearSynchronization is responsible for sending messages to the android app  */
class AndroidSynchronization: WearSynchronizationInterface() {

    @Inject(parameters = ["AndroidSynchronization"]) private lateinit var logger: Logger
    private val messageClient = Wearable.getMessageClient(RPout.getAppContext())
    private val capabilityClient = Wearable.getCapabilityClient(RPout.getAppContext())

    /**
     * Returns a list of nodes that are available for message sending through the provided callback.
     * Errors are silently dropped and only written to the log file
     */
    fun getNodes(callback: (Set<Node>) -> Unit) {
        val nodes = capabilityClient.getCapability("android", CapabilityClient.FILTER_REACHABLE)
        nodes.addOnSuccessListener {
            callback(it.nodes)
        }
        nodes.addOnFailureListener {
            logger.log("ee", "Failed to get nodes from capability client: ${it.message}")
        }
    }

    fun sendTextMessage(type: MessageType, message: String, onlyNearby: Boolean = false, onSuccess: () -> Unit) {
        // Require at least one node
        getNodes {
            if (it.isEmpty()) {
                logger.log("w", "No connected devices available for sending message of type: " + type.name)
            } else {
                // Send message to all devices
                it.forEach { node ->
                    if (!onlyNearby || node.isNearby) {
                        val task = messageClient.sendMessage(node.id, type.path, message.toByteArray())

                        task.addOnCompleteListener { result ->
                            if (result.isSuccessful) {
                                Log.i("RPout-Logger", "Sent message of type " + type.name + " successfully")
                                onSuccess()
                            } else {
                                logger.log("w", "Failed to send message of type " + type.name)
                            }
                        }
                    }
                }
            }
        }
    }

    override fun sendTextMessage(type: MessageType, message: String, onSuccess: () -> Unit) {
        sendTextMessage(type, message, false, onSuccess)
    }


}