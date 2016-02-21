# BT Cloud Compute driver for Docker Machine

## Status
_*** WORK IN PROGRESS ***_

This package is incomplete.


## Create a machine

```
$ docker-machine create -d btcloud <vm name>
Running pre-create checks...
Creating machine...
Waiting for machine to be running, this may take a few minutes...
Detecting operating system of created instance...
Waiting for SSH to be available...
Detecting operating system of created instance...
Provisioning created instance...
Copying certs to the local machine directory...
Copying certs to the remote machine...
Setting Docker configuration on the remote daemon...
To see how to connect Docker to this machine, run: docker-machine env example
```

Options:

| Option                      | Environment Variable      | Description                           |Required |
|-----------------------------|:-------------------------:|---------------------------------------|--------:|
| --btcloud-endpoint          | BTCLOUD_ENDPOINT          | BT Cloud Compute API endpoint         | N |
| --btcloud-api-key           | BTCLOUD_API_KEY           | BT Cloud Compute API key              | Y |
| --btcloud-secret-key        | BTCLOUD_SECRET_KEY        | BT Cloud Compute secret key           | Y |
| --btcloud-ssl               | BTCLOUD_SSL               | Verify BT Cloud Compute SSL           | N |
| --btcloud-zone              | BTCLOUD_ZONE              | BT Cloud Compute availability zone    | Y |
| --btcloud-template          | BTCLOUD_TEMPLATE          | BT Cloud Compute template             | Y |
| --btcloud-service-offering  | BTCLOUD_SERVICE_OFFERING  | BT Cloud Compute service offering     | Y |

##Acknowledgements
This package uses the excellent [go-cloudstack] (https://github.com/xanzy/go-cloudstack) API client by Sander van Harmelen (<sander@xanzy.io>)

## Author
Christian Lafferty, BT Research & Innovation.

##License
Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at <http://www.apache.org/licenses/LICENSE-2.0>

Copyright (c) 2016 BT Group plc
