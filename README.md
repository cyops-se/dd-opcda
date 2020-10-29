# dd-opcda - Simple OPC DA data one way replicator
This very basic program collects all tags from an OPC DA source once a second and sends them in JSON format over UDP to a single-cast receiver IP.

The primary use of this program is for replicating real time data from sensitive systems through a data diode to a potentially hostile network in order to achieve complete isolation from the external network.