#!/bin/bash

find .|grep md$|xargs -n1 tocenize -max 3 -min 2
