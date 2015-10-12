(Yet another) Cool compiler
===========================

**Note:** This is not [the `coolc` that outputs BIT](https://github.com/BenLubar/bit/tree/master/cmd/coolc). This `coolc` outputs 32-bit x86 assembly with AT&T (GNU as) syntax.

Calling convention
------------------

- The reciever is pushed onto the stack first, followed by the arguments in the same order that they are listed in the source code. That is, 8(%ebp) is the last argument, 12(%ebp) is the second-last, and so on.
- The return value is in the AX register.
