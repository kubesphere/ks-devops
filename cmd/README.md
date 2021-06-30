There're two components in `ks-devops`, they are APIServer and controller-manager.
Normally, they run in a Kubernetes cluster as Pods. But technically, they also are
regular executable binary files. So, you can run `ks-devops` as a binary file.

There're three commands here:

* [apiserver](apiserver)
* [controller-manager](controller)
* [All in One](allinone)
    * Combine apiserver and controller-manager into one command.
