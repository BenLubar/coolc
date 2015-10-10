package main

var basicCool = []byte("\n\n" + `// Basic classes for Cool 2011
// John Boyland
// January 2011

// This file includes classes which are treated specially by the compiler.
// Native features are implemented in the Cool runtime system,
// and are only permitted for features defined in this file.

// Classes with native attributes 
// (Unit, Int, Boolean, String, Symbol and ArrayAny)
// may not be inherited from.

/** The Any class is the root of the inheritance hierarchy. */
class Any() extends native {

  /** Returns a string representation for the object */
  def toString() : String = native;

  /** return true if this object is equal (in some sense) to the argument */
  def equals(x : Any) : Boolean = native;
}

/** The IO class provides simple input and output operations */
class IO() {

  /** Terminates program with given message. 
   * Return type of native means that 
   * (1) result type is compatible with anything
   * (2) function will not return.
   */
  def abort(message : String) : Nothing = native;

  /** Print the argument (without quotes) to stdout and return itself */
  def out(arg : String) : IO = native;

  def is_null(arg : Any) : Boolean = {
    arg match { 
      case null => true 
      case x:Any => false 
    }
  };

  /** Convert to a string and print */
  def out_any(arg : Any) : IO = {
    out(if (is_null(arg)) "null" else arg.toString())
  };

  /** Read and return characters from stdin to the next newline character. 
   * Return null on end of file.
   */
  def in() : String = native;

  /** Get the symbol for this string, creating a new one if needed. */
  def symbol(name : String) : Symbol = native;

  /** Return the string associated with this symbol. */
  def symbol_name(sym : Symbol) : String = native;
}

/** A class with no subclasses and which has only one instance.
 * It cannot be instantiated of inherited.
 * The null pointer is not legal for Unit.
 */
class Unit() { }

/** The class of integers in the range -2^31 .. (2^31)-1
 * null is not a legal value for integers, and Int can have no subclasses.
 */
class Int() {
  var value = native;

  /** Convert to a string representation */
  override def toString() : String =
    if (this < 0)
      "-".concat({
        var n : Int = -this;
        if (n < 0)
          "214743648"
        else
          n.toString()
      })
    else {
      var digits : String = "0123456789";
      var s : String = "";
      var n : Int = this;
      while (0 < n) {
        var n10 : Int = n / 10;
        var d : Int = n - n10 * 10;
        s = digits.substring(d, d + 1).concat(s);
        n = n10
      };
      if (s.length() == 0)
        "0"
      else
        s
    };

  /** Return true if the argument is an int with the same value */
  override def equals(other : Any) : Boolean = native;
}

/** The class of booleans with two legal values: true and false.
 * null is not a legal boolean.
 * It is illegal to inherit from Boolean.
 */
class Boolean() {
  var value = native;

  /** Convert to a string representation */
  override def toString() : String = if (this) "true" else "false";
}

/** The class of strings: fixed sequences of characters.
 * Unlike previous version of Cool, strings may be null.
 * It is illegal to inherit from String.
 */
class String() {
  var length : Int = 0;
  var str_field = native;

  override def toString() : String = this;

  /** Return true if the argument is a string with the same characters. */
  override def equals(other : Any) : Boolean = native;

  /** Return length of string. */
  def length() : Int = length;

  /** Return (new) string formed by concatenating self with the argument */
  def concat(arg : String) : String = native;

  /** returns the  substring of self beginning at position start
   * to position end (exclusive)
   * A runtime error is generated if the specified
   * substring is out of range.
   */
  def substring(start : Int, end : Int) : String = native;

  /**
   * Return the character at the given index of the string
   * as an integer.
   */
  def charAt(index : Int) : Int = native;

  /**
   * Return the first index of given substring in this string,
   * or -1 if no such substring.
   */
  def indexOf(sub : String) : Int = {
    // we give a default implementation that is wasteful of space
    // but which enables us to use Cool to write it:
    var n : Int = sub.length();
    var diff : Int = length - n;
    var i : Int = 0;
    var result : Int = -1;
    while (i <= diff) {
      if (substring(i,i+n) == sub) {
	result = i;
	i = diff + 1
      } else {
	i = i + 1
      }
    };
    result
  };
}

/**
 * A symbol is an interned string---two symbols with the same string
 * are always identically the same object.  This has two benefits: <ol>
 * <li> equality checking is very efficient
 * <li> we can have a fixed hash code for each symbol. </ol>
 * Creating symbols is restricted to ensure the uniqueness properties.
 * See IO.symbol.  In "Extended Cool", the name is immutable and
 * can be accessed directly.  In Cool, we use IO.symbol_name.
 */
class Symbol() {
  var next : Symbol = null;
  var name : String = "";
  var hash : Int = 0;

  override def toString() : String = "'".concat(name);

  def hashCode() : Int = hash;
}

/** An array is a mutable fixed-size container holding any objects.
 * The elements are numbered from 0 to size-1.
 * An array may be void.  It is not legal to inherit from ArrayAny.
 */
class ArrayAny(var length : Int) {

  var array_field = native;

  /** Return length of array. */
  def length() : Int = length;

  /** Return a new array of size s (the original array is unchanged).  
   * Any values in the original array that fit within the new array 
   * are copied over.  If the new array is larger than the original array,
   * the additional entries start void.  If the new array is smaller 
   * than the original array, entries past the end of the new array are 
   * not copied over.
   */
  def resize(s : Int) : ArrayAny = {
    var a : ArrayAny = new ArrayAny(s);
    var i : Int = 0;
    while (if (i < length) i < s else false) {
      a.set(i, get(i));
      i = i + 1
    };
    a
  };

  /* Returns the entry at location index.
   * precondition: 0 <= index < length()
   */
  def get(index : Int) : Any = native;

  /* change the entry at location index.
   * return the old value, if any (or null).
   * precondition: 0 <= index < length()
   */
  def set(index : Int, obj : Any) : Any = native;
}
`)
