<h1 align="center">Ontology Layer2 Node</h1>
<h4 align="center">Version 1.0.0 </h4>

[中文](README_CN.md) | English

The Ontology Layer2 is aimed at creating a solution to off-chain scaling. In order to cater to the needs of users that demand low transaction fee and low latency, the layer2 system can help address the scaling needs of complex applications.

## Node Installation

### Pre-requisites

- Golang - v1.14 or higher
- Suitable Go development environment
- Linux OS

### Fetching Layer2 Node Source Code

Clone the Layer2 repo to the `$GOPATH/src/github.com/ontio` directory by running the following shell command:

```shell
$ git clone https://github.com/ontio/layer2.git
```

### Compilation

Run the `make` command to compile the library.

```shell
$ make all
```

Upon successful compilation, an executable program will appear in the directory.

- `Node`: The node program used to control the node operations using the command line.

### Running the Layer Node

Start the node by running the following command:

``` shell
./Node
```

## License

The Ontology library is licensed under the GNU Lesser General Public License v3.0, read the LICENSE file in the root directory of the project for details.
