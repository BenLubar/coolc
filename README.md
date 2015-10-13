(Yet another) Cool compiler
===========================

[![Build Status](https://drone.io/github.com/BenLubar/coolc/status.png)](https://drone.io/github.com/BenLubar/coolc/latest)

**Note:** This is not [the `coolc` that outputs BIT](https://github.com/BenLubar/bit/tree/master/cmd/coolc). This `coolc` outputs 32-bit x86 assembly with AT&T (GNU as) syntax.

Calling convention
------------------

- The reciever is pushed onto the stack first, followed by the arguments in the same order that they are listed in the source code. That is, `8(%ebp)` is the last argument, `12(%ebp)` is the second-last, and so on.
- The return value is in the AX register.

Memory layout
-------------

- Heap consists of objects end to end.
- Objects have a tag, a size, and a garbage collector tag, followed by the data of the object. The 3-word header is not included in the size.
- The tag is either `-1` for garbage, `0` for the end of the heap, or a number from `1` to `max_tag`, inclusive, for a class type.
- Each positive tag number has an associated method table, name, and pointer coount.
- GC tags are negative for certain special cases like permanent objects and garbage, and otherwise contain the number of stack references to the object. Non-garbage objects with a non-zero GC tag are considered roots of the heap.
