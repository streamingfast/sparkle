#!/bin/bash

cat reqs.txt | vegeta attack -rate=500 -duration=30s | tee results.bin | vegeta report

cat results.bin | vegeta plot > plot.html
