MedTech Chain Cyber Threat Intelligence (MedCTI)
================================================

Description
-----------
This project contains a sample implementation to show how privacy-preserving CTI can be implemented on the MedTech Chain. The project is split into several implementations each showing a different aspect of the solution.

The implementations have a consumer and provider variant. The service consumer owns the data and is able to view analytical results. The service provider uses data of the consumers to provide analytical results.

In a setting of cyber threat intelligence on medical devices the hospital is a consumer and the threat hunter is a provider.

Implementations
---------------

* plaintextdemo: This folder contains an implementation that shows what components exists and how they cooperate with each other. All messages are visible in plain text. Cryptographic functions are simulated, but not implemented.