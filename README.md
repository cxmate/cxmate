cxMate
======

[![CircleCI](https://circleci.com/gh/cxmate/cxmate.svg?style=svg)](https://circleci.com/gh/cxmate/cxmate)
[![Test Coverage](https://codeclimate.com/github/cxmate/cxmate/badges/coverage.svg)](https://codeclimate.com/github/cxmate/cxmate)
[![Issue Count](https://codeclimate.com/github/cxmate/cxmate/badges/issue_count.svg)](https://codeclimate.com/github/cxmate/cxmate)
[![Go Report Card](https://goreportcard.com/badge/github.com/cxmate/cxmate)](https://goreportcard.com/report/github.com/cxmate/cxmate)

<img align="right" height="300" src="http://www.cytoscape.org/images/logo/cy3logoOrange.svg">

---

cxMate streams Cytoscape networks directly to your code. cxMate listens for CX on a network, Cytoscape's native network format, and transforms it data structures in your's codes native language. cxMate will any networks you want to send to Cytoscape back into CX. When you use cxMate, your code can also be used by any other tool, service, or program that speaks CX like an NDEx server. The requests that cxMate responds to are plain HTTP calls, so your service can also be called from any HTTP client, such as the requests Python module from a Jupyter notebook, or curl.

---

_cxMate is an official [Cytoscape](http://www.cytoscape.org) project written by the Cytoscape team._

Installation
------------

cxMate comes precompiled as a static binary for a number of platforms, and exists on Docker Hub as a container image that can be run with Docker, Docker Swarm, or Kubernetes.

- Download a precompiled static binary for your platform [here](https://github.com/ericsage/cxmate/releases)
- Run cxmate in a docker container with the [official Docker Hub image](https://hub.docker.com/r/ericsage/cxmate/)

Getting Started
---------------

The easiest way to use cxMate is with an official cxMate SDK. The cxMate SDKs provide easy to understand boilerplate and adapters to popular network formats. Each SDK has it's own Getting Started readme for creating a service in that language. Once you've implemented your cxMate service, you'll also need to write a `cxmate.json` file to configure cxMate's runtime behavior and give cxMate information about your service. See the Configuration section for details.

Offical cxMate SDKs:
- [Python](http://github.com/cxmate/cxmate-py)

Configuration
-------------

On startup, cxMate will look for a `cxmate.json` file in its current directory. This file contains a JSON configuration object that cxMate uses to interact with your service. cxMate will not start if this file is missing or required fields are not provided.

Example:
```json
{
  "general": {
    "location": "0.0.0.0:80",
    "domain": "echo.cytoscape.io",
    "logger": {
      "debug": true,
      "file": "echo.log",
      "format": "json"
    }
  },
  "service": {
    "location": "localhost:8080",
    "title": "echo",
    "author": "Eric Sage",
    "email": "edsage@ucsd.edu",
    "description": "A test service that echos its input to its output.",
    "website": "http://github.com/ericsage/echo",
    "repository": "http://github.com/ericsage/echo",
    "license": "MIT",
    "language": "Python",
    "parameters": [
      {
        "key": "test_param",
        "default": "1.0",
        "description": "A parameter may be any string encoded value. The default value is garunteed to reach the service."
      }
    ],
    "input": [
      {
        "label": "Input",
        "description": "An input network to be echoed",
        "aspects": ["nodes", "edges", "nodeAttributes", "edgeAttributes", "networkAttributes"]
      }
    ],
    "output": [
      {
        "name": "Output",
        "description": "An output network which is the same network as the input.",
        "aspects": ["node", "edge", "nodeAttribute", "edgeAttribute", "networkAttribute"]
      }
    ]
  }
}
```

#### General
Configures how cxMate will operate as a service proxy.

| Option   | Required | Default | Description                                                                                                           |
| -------- | -------- | ------- | --------------------------------------------------------------------------------------------------------------------- |
| location | true     | N/A     | The address and port cxMate will listen for requests on, e.g. "0.0.0.0:80" will listen on all interfaces on port 80.  |
| domain   | false    | ""      | The HTTP URL cxMate will listen on.                                                                                   |
| logger   | false    | N/A     | See Logger                                                                                                            |

#### Logger
Configures how the cxMate standard logger will operate. By default the logger will output text to stdout without debugging information.

| Option | Required | Default | Description                                                                             |
| ------ | -------- | ------- | --------------------------------------------------------------------------------------- |
| debug  | false    | false   | Logs extra debugging information when set.                                              |
| file   | false    | ""      | If set, the logger logs to the speicifed file (creating the file if it does not exist). |
| format | false    | ""      | Sets the format of the log messages. Supported values are 'text' and 'json'.            |

#### Service
Configures how cxMate will interact with the backing service, and also provides service metadata to cxMate.

| Option          | Required | Default | Description                                                                                                                                 |
| --------------- | -------- | ------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| location        | true     | N/A     | The address and port cxMate will contact the service on.                                                                                    |
| title           | true     | N/A     | The title of the service.                                                                                                                   |
| version         | true     | N/A     | A SemVer version identifer for the service, e.g. "2.0.0", "1.0.0-alpha".                                                                    |
| author          | false    | ""      | The name of the primary author of the service.                                                                                              |
| email           | false    | ""      | The email of the primary author of the service.                                                                                             |
| description     | false    | ""      | A brief description of what the service does.                                                                                               |
| website         | false    | ""      | A user facing website that service users can visit to learn more about the service.                                                         |
| repository      | false    | ""      | The source code repository of the service.                                                                                                  |
| language        | false    | ""      | The programming language the service is primarily written in.                                                                               |
| parameters      | false    | []      | A list of Parameters. See Parameter.                                                                                                        |
| input           | true     | N/A     | A list of NetworkDescriptions. Each NetworkDescription describes a network expected as input. See NetworkDescription.                       |
| singletonInput  | false    | false   | If set, only the first element in the array of inputs will be used. It will be expected as a singleton network in the input of the service. |
| output          | true     | N/A     | A list of NetworkDescriptions. Each NetworkDescription describes a network sent as output. See NetworkDescription.                          |
| singletonOutput | false    | false   | If set, only the first element in the array of outputs will be used. It will be sent as a singleton network in the output of the service.   |

#### Parameter
A parameter expected by the service. cxMate garuntees that the service will receive at least one parameter for each parameter defined (default values are required), however, multiple query string parameters with the key name will send multiple cxMate parameters to the service.

| Option      | Required | Default  | Description                                                                             |
| ----------- | -------- | -------- | --------------------------------------------------------------------------------------- |
| name        | true     | N/A      | The name of the parameter will be matched against the query string parameters sent to the service, e.g. "?heat=1.0" will match name "heat".                                                                                |
| default     | true     | N/A      | The default value of the paramater. If no query string parameter matches this parameter, the default value will be sent.                                                                                                   |
| description | false    | ""       | A brief description of what the parameter represents and its purpose in the service.
| type        | false    | "string" | The type cxMate will cast the query string to before sending the parameter to the service. Must be one of "number", "integer", "boolean", or "string", e.g. if type is set to number "?match=1.0"  will cast "1.0" to 1.0. |
| format      | false    | ""       | Extra semantic information, e.g. "password", "float64", "secret", "gene", etc.                                                                                                                                             |

#### NetworkDescription
Describes a CX network.

| Option      | Required | Default  | Description                                                                             |
| ----------- | -------- | -------- | --------------------------------------------------------------------------------------- |
| label       | true     | N/A      | The name of the parameter will be matched against the query string parameters sent to the service, e.g. "?heat=1.0" will match name "heat".                                                                                |
| description | false    | ""       | A brief description of what the network represents and its purpose in the service.
| aspects     | true     | N/A      | The CX aspects that will appear in the network. For an Input network, the network must contain *at least* these aspects. Any aspects not in this list will not be forwarded to the service. For an Output Network, these are all the aspects taht *can* appear in the network. |

Contributors
------------

We welcome all contributions via Github pull requests. We also encourage the filing of bugs and features requests via the Github [issue tracker](https://github.com/cxmate/cxmate/issues/new). For general questions please [send us an email](eric.david.sage@gmail.com).

License
-------

cxmate is MIT licensed and a product of the [Cytoscape Consortium](http://www.cytoscapeconsortium.org).

Please see the [License](https://github.com/cxmate/cxmate/blob/master/LICENSE) file for details.
