# proc

Process control, logging and child processes

## Badges

<insert badges here

## Description

Golang is a great language for concurrency, but sometimes you want parallelism,
and it's often simpler to design apps as little units that plug together via 
RPC APIs. Services don't always have easy start/stop/reconfig/restart 
controls, and sometimes they use their own custom configuration scheme that is
complicated to integrate with yours.

In addition, you may want to design your applications as neat little pieces, but
how to attach them together and orchestrate them starting up and coming down 
cleanly, and not have to deal with several different ttys to get the logs.

Proc creates a simple framework for creating observable, controllable execution
units out of your code, and more easily integrating the code of others.

This project is the merge of several libraries for logging, spawning and
controlling child processes, and creating an RPC to child processes that 
controls the run of the child process. Due to the confusing duplication of 
signals as a means of control and the lack of uniformity in signals (ie 
windows) the goal of `proc` is to create one way to these things, following 
the principles of design used for Go itself.

Initially I was reworking this up from past code but decided that the job 
will now be about taking everything from [p9](https://github.com/cybriq/p9), 
throwing away the stupid struct definition/access method (it just 
complicates things) because that implementation has almost everything in it. 
There is many things inside the `pod` subfolder on that repository that need 
to be separated. It was my first project and learning how to structure and 
architect was a big challenge considering prior to that I had only done my 
main work in algorithms.

The `p9` work includes almost everything, and discards some design concepts 
that were used in previously used libraries like `urfave/cli` which 
unnecessarily complicated the scheme by putting configurations underneath 
commands, unnecessary because a program with many run modes still only has 
one configuration set and forcing the developer to think of them as a 
unified set instead of children of commands saves a lot of complexity and 
maintenance.

We may add some more features, such as an init function to each 
configuration item, and eliminate the complexity of initial startup code to 
put things together that are conceptually and functionally linked.

## Installation

### For developers:

To make nice tagged versions with the version number in the commit as well as
matching to the tag, there is a tool called [bumper](cmd/bumper) that bumps 
the version to the next patch version (vX.X.PATCH), embeds this new version
number into [version.go](./version.go) with the matching git commit hash of the
parent commit. This will make importing this library at an exact commit much 
more human.

In addition, it makes it easy to make a nice multi line commit message as many
repositories request in their CONTRIBUTION file by replacing ` -- ` with two 
carriage returns.

To install:

    go install ./bumper/.

To use:

    bumper make a commit comment here -- text after double \
        hyphen will be separated by a carriage return -- \
        anywhere in the text

To automatically bump the minor version use `minor` as the first word of the
comment. For the major version `major`.

## Usage

## Support

## Contributing

## Authors and acknowledgment

David Vennik david@cybriq.systems

## License

Unlicensed: see [here](./LICENSE)

## Project status

In the process of revision and merging together several related libraries that
need to be unified.