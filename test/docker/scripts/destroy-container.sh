#!/bin/bash

function destroy_container() {
  name=$1
  docker stop "migrator-$name"
  docker rm "migrator-$name"
}
