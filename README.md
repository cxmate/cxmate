cxmate
======

cxmate is a RESTful network API proxy service for network algorithms. If you're interested in turning a network algorithm into a robust web service, cxmate can drastically reduce the investment of time and effort required by providing the following key features:

- **Streaming support for CX, an extensible aspect oriented network network interchange format**<br>
  CX supports steaming of arbitrarily large networks, and is well suited for encoding rich networks through the use of aspects. cxmate reads and writes streams of CX, allowing high throughput with lower memory consumption. Your algorithm need not know the exact details of CX to take advantage of its power and flexibility. cxmate supports one-to-one, one-to-many, and many-to-many network algorithms. You decide how many networks cxmate will receive and send.
  
- **Work with native objects in your algorithm's language instead of dealing with HTTP request and responses**<br>
  cxmate provides an efficient translation between the CX interchange format and objects native to your algorithm. By the time cxmate calls your algorithm, your code will receive a stream of easy to use element objects containing network pieces, algorithm parameters, and formatted errors to work with. cxmate only expects a stream of native objects in return. Never work with raw HTTP again.
  
- **A fully RESTful JSON HTTP interface managed by cxmate on behalf of your service**<br>
  Any algorithm proxied by cxmate need not worry about writing HTTP handlers, URL parsers, or dealing with any of the boilerplate associated with creating a RESTful web service. Your clients will have full access to the popular REST method of interfacing with your algorithm through cxmate, allowing you to focus on writing and maintaining algorithm logic instead of service interfaces.
  
- **Algorithm parameters and error handling made easy**<br>
  When cxmate receives a request, query string parameters are automatically translated to key/value elements and streamed to your algorithm like any other object. Any errors detected by cxmate while parsing the incoming network and parameters will also be turned into error objects your algorithm can then decide to send back to the client, handle internally, or ignore.
  
- **Service insights via automated metrics gathering and logging**<br>
  cxmate exposes a plethora of useful statistics about itself and the proxied service via its RESTful HTTP API, allowing service authors to monitor the health and usage of their service over time.  
 
 cxmate is a subproject of Cytoscape and the Ideker Lab at the University of California, San Diego. cxmate greatly decreases the time bioinformaticians, computer scientists, and researchers from other disciplines spend writing code, allowing them to focus on their algorithms and providing biological value to research community. cxmate also decreases the time spent creating services for features used by tens of thousands of Cytoscape users every day.

Installation
------------

While we recommend eventually running cxmate and your service in Docker containers for maximum portability and deployability, we also precompile cxmate binaries for popular platforms for testing and development:

- Download a precompiled binary for your platform [here](https://github.com/ericsage/cxmate/releases)
- Run cxmate in a docker container with the [official Docker Hub image](https://hub.docker.com/r/ericsage/cxmate/)

Python Getting Started
----------------------

An offical cxMate module for Python exists on PyPi that makes developing Python cxMate services easy.

```python
from cxmate.service inport Service, Stream

class MyEchoService(Service):
    """
    MyService is a subclass 
    """
    
    def process(self, input_stream):
        """
        process is a required method, if it's not implemented, cxmate.service will throw an error
        this process implementation will echo the received network back to the sender
        
        :param input_stream: a python generator that returns CX elements
        :returns: a python generator that returns CX elements
        """
        network = Stream.to_networkx(input_stream)
        return Stream.from_networkx(network)
        
if __name__ == "__main__":
  myService = MyService()
  myService.run() #run starts the service listening for requests from cxMate
```

Configuration
-------------
On startup, cxMate will look for a `cxmate.json` file to load configuraiton from. cxMate will not start if it cannot load this file.

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
    "name": "echo",
    "author": "Eric Sage",
    "summary": "A test service that echos its input to its output.",
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

Contributors
------------

We welcome all contributions via Github pull requests. We also encourage the filing of bugs and features requests via the Github [issue tracker](https://github.com/ericsage/cxmate/issues/new). For general questions please [send us an email](eric.david.sage@gmail.com).

License
-------

cxmate is MIT licensed and a product of the [Cytoscape Consortium](http://www.cytoscapeconsortium.org).

Please see the [License](https://github.com/ericsage/cxmate/blob/master/LICENSE) file for details.
