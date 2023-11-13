# Stack

[![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/rs/zerolog/master/LICENSE)

Stack provides stack tracing and enhanced error logging. Much of this code predates `pkg/errors` supporting marshallable stacktraces but has been updated to have interoperability with it.

## Features
* Grabs full traces at the invocation of the error trace
* Prevents repeated nesting of traces with automatic merging of new error data
* Allows for additional data to be attached to errors that can be optionally exported based on the error handling
* Provides automatic marshalling of entries to Zerolog