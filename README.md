#Ninja Sphere Go Zwave Driver

##Building
To perform an isolated build using only the current source and dependencies pulled from github, Run `make` in the directory of the driver.

To perform a development build using dependencies pulled from siblings of the current directory, run `make here`.

Note that because this driver depends on a C++ component, a cross-compile on OSX to produce the linux/arm executeble is not possible.
To build the linux/arm target, execute the build natively on a linux/arm host.

##Date
2014-09-19