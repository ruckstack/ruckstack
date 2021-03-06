# Ruckstack

**The modern application server**

Ruckstack packages your entire application stack into a easily self-contained, easily manged installation.

For more information, see http://ruckstack.org

## Developing Ruckstack

### Building

Ruckstack can be built with `./BUILD.sh`. 

This will collect and generate the bindata data, compile everything, and build the distribution archives into the `out` directory.  

All generated and compiled files can be cleared out with `BUILD.sh clean`

### Code Structure

The `cmd` directory contains the source for the command line interfaces. There are separate sub-directories for each CLI we have.
Most of the actual logic for these CLIs live in the `internal` directory, which also has subdirectories for each section we have.

The `dist` directory contains files that will be bundled into the ruckstack tgz/zip

All compiled code goes to the `out` directory 

### Development Environment

#### GO

Ruckstack is written in [Go](http://golang.org)

#### Linux

Because [k3s](http://k3s.io) can only run on Linux, Ruckstack development works best on Linux

#### Ruckstack Builder

When working on the Ruckstack Builder, it is fastest to compile and run the ruckstack binary on its own. 

#### System-Control

System-Control is the internal name for the CLI that runs and manages the installed system. It will be renamed during the packaging process based on project configuration.
   
When working on system-control, it is fastest to compile and run the system-control binary on its own. 

Because system-control runs as part of a server installation, you will need to have SOME version of a ruckstack-built system installed locally.
But, by running the built system-control binary with a `RUCKSTACK_HOME=/path/to/local/install` environment variable, your development system-control binary will 
run as if its ran as part of that local installation.

## License

Ruckstack is licensed under the Apache 2.0 license.

