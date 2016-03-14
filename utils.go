package main

func left(a, _ interface{}) interface{} {
  return a
}

func right(_, b interface{}) interface{} {
  return b
}

func inRange(x, l, r int) bool {
  if x < l || r < x {
    return false
  }
  return true
}
