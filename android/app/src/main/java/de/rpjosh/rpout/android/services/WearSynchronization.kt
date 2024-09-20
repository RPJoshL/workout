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
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.runBlocking
import java.util.concurrent.Flow

/** WearSynchronization is responsible for sending messages to the Wearable app  */
class WearSynchronization: WearSynchronizationInterface() {

    @Inject(parameters = ["WearSynchronization"]) lateinit var logger: Logger
    private val messageClient = Wearable.getMessageClient(RPout.getAppContext())
    private val capabilityClient = Wearable.getCapabilityClient(RPout.getAppContext())

    /**
     * Returns a list of nodes that are available for message sending through the provided callback.
     * Errors are silently dropped and only written to the log file
     */
    fun getNodes(callback: (Set<Node>) -> Unit) {
        val nodes = capabilityClient.getCapability("wear", CapabilityClient.FILTER_REACHABLE)
        nodes.addOnSuccessListener {
            callback(it.nodes)
        }
        nodes.addOnFailureListener {
            logger.log("ee", "Failed to get nodes from capability client: ${it.message}")
        }
    }

    fun addCapability(name: String) {
        capabilityClient.addLocalCapability(name)
    }

    override fun sendTextMessage(type: MessageType, message: String, onSuccess: () -> Unit) {
        // Require at least one node
        getNodes {
            if (it.isEmpty()) {
                responseViewInterface.displayError(Tr.get("sync_noConnectedDevices"))
            } else {
                // Send message to all devices
                it.forEach { node ->
                    val task = messageClient.sendMessage(node.id, type.path, message.toByteArray())

                    task.addOnCompleteListener { result ->
                        if (result.isSuccessful) {
                            responseViewInterface.displaySuccess(Tr.get("sync_successfully"))
                            onSuccess()
                        } else {
                            responseViewInterface.displayError(Tr.get("sync_error"))
                        }
                    }
                }
            }
        }
    }


}

