// fizzbuzz.js — sample file for the read / edit / search tool demos.
// Prints 1..100, replacing multiples of 3 with "Fizz", multiples of 5 with
// "Buzz", and multiples of both with "FizzBuzz".

function fizzbuzz(n) {
  for (let i = 1; i <= n; i++) {
    if (i % 15 === 0) {
      console.log("FizzBuzz");
    } else if (i % 3 === 0) {
      console.log("Fizz");
    } else if (i % 5 === 0) {
      console.log("Buzz");
    } else {
      console.log(i);
    }
  }
}

fizzbuzz(100);
