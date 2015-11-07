package main

var coroutineCool = []byte("\n\n" + `// Coroutine support in Cool.

// This is an interface for Communicating Sequential Processes (CSP) in Cool.
// https://en.wikipedia.org/wiki/Communicating_sequential_processes
// It is heavily influenced by Go's goroutines and channels.

class Runnable() extends IO() {
	def run() : Unit = abort("Runnable.run is abstract in ".concat(super.toString()));
}

// A Coroutine runs the Runnable given to it concurrently with all other
// Coroutines. The order of any action done without synchronization is
// unspecified. If a Coroutine does not give up control (by terminating or
// doing a Channel operation), other Coroutines may never run.
//
// new Main() is called by a coroutine. When Main's constructor returns,
// the program terminates, regardless of the state of any other Coroutine.
class Coroutine(var runnable : Runnable) {
	var coroutine_field = native;
}

// Objects can be passed through Channels. A Coroutine that calls send or recv
// will wait until another coroutine calls recv or send, respectively. If
// multiple Coroutines are sending or receiving, the order they are handled
// is unspecified.
class Channel() {
	var channel_field = native;

	def send(x : Any) : Unit = native;
	def recv() : Any = native;
}
`)
