#!/bin/bash
grep -zoP '```package\s*\K[\s\S]*(?=```)'
