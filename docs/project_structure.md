# /configs
In this directory we have structures for storing data of the raw config loaded from the `config.yaml`. After loading ended we set undefined values with default ones. They are needed for a short time until context has been initialised.
# /context
Here we have types represents global state of the Project and each Application. They are initialized by data from configs. Context values live the whole software lifecycle (until the terminating software).
# /internal
This directory consist packages used by all parts of software. They can't be imported by another software.
## /collections
Structures for some abstract collections. It very common and can probably describe any collection. It can be used by any plugin for their needs to describe some collection in a storage.
## /plugins
This directory contains declaration of plugin types. Also it has helper functions and structs. These packages should be imported in packages which implement certain plugins. (More details about plugin types in `plugins.md`)
# /jsonpath
In this package we have method that can extract some data from the json by the given json path directive (look for json path syntax). 

https://goessner.net/articles/JsonPath/ 

https://support.smartbear.com/alertsite/docs/monitors/api/endpoint/jsonpath.html
# /plugins
This directory contains certain implementations of plugins. They are fulfill plugin declarations.