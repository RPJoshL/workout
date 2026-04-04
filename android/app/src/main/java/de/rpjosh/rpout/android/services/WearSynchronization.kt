package de.rpjosh.rpout.android.services

import com.google.android.gms.wearable.CapabilityClient
import com.google.android.gms.wearable.Node
import com.google.android.gms.wearable.Wearable
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.services.WearSynchronizationInterface
import kotlin.collections.forEach

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
            callback(emptySet())
            logger.log("i", "Failed to get nodes from capability client: ${it.message}")
        }
    }

    fun addCapability(name: String) {
        capabilityClient.addLocalCapability(name)
    }

    override fun showNotConnectedMessage() {
        responseViewInterface.displayError(Tr.get("sync_noConnectedDevices"))
    }

    override fun sendTextMessage(type: MessageType, message: String, onError: () -> Unit, onSuccess: () -> Unit) {
        // Require at least one node
        getNodes {
            if (it.isEmpty()) {
                onError()
            } else {
                // Send message to all devices
                sendTextMessageToNodes(it, type, message, onSuccess)
            }
        }
    }

    fun sendTextMessageToNodes(nodes: Set<Node>, type: MessageType, message: String, onSuccess: () -> Unit) {
        nodes.forEach { node ->
            val task = messageClient.sendMessage(node.id, type.path, message.toByteArray())

            task.addOnCompleteListener { result ->
                if (result.isSuccessful) {
                    onSuccess()
                } else {
                    responseViewInterface.displayError(Tr.get("sync_error"))
                }
            }
        }
    }


}

