#!/usr/bin/env python3
# Copyright (c) 2020, The OTNS Authors.
# All rights reserved.
#
# Redistribution and use in source and binary forms, with or without
# modification, are permitted provided that the following conditions are met:
# 1. Redistributions of source code must retain the above copyright
#    notice, this list of conditions and the following disclaimer.
# 2. Redistributions in binary form must reproduce the above copyright
#    notice, this list of conditions and the following disclaimer in the
#    documentation and/or other materials provided with the distribution.
# 3. Neither the name of the copyright holder nor the
#    names of its contributors may be used to endorse or promote products
#    derived from this software without specific prior written permission.
#
# THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
# AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
# IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE
# ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE
# LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR
# CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF
# SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS
# INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN
# CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE)
# ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE
# POSSIBILITY OF SUCH DAMAGE.

import time

from otns.cli import OTNS

XGAP = 70
YGAP = 70
RADIO_RANGE = int(XGAP*1.5)


def main():
    ns = OTNS(otns_args=['-log', 'debug'])
    ns.web()
    ns.speed = float('inf')

    while True:
        # wait until next time
        test_nxn(ns, 8)


def test_nxn(ns, n):
    nodes = ns.nodes()
    for id in nodes:
        ns.delete(id)

    ns.countdown(n * n, f"Testing {n}x{n} nodes ... %v seconds left")
    for r in range(n):
        for c in range(n):
            ns.add("router", 50 + XGAP * c, 50 + YGAP * r, radio_range=RADIO_RANGE)

    secs = 0
    while secs < 1800:
        ns.go(100)
        secs += 100

        partitions = ns.partitions()
        if len(partitions) == 1 and 0 not in partitions:
            # all nodes converged into one partition
            break

    ns.go(600)

if __name__ == '__main__':
    main()
