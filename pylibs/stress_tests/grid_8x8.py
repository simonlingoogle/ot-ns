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
import random
import unittest

from StressTestBase import StressTestBase

XGAP = 70
YGAP = 70
RADIO_RANGE = int(XGAP * 1.5)

PING_INTERVAL = 10
PING_COUNT = 1


class BasicTests(StressTestBase):

    def test(self):
        ns = self.ns
        ns.web()

        while True:
            # wait until next time
            self.test_nxn(ns, 8)

    def test_nxn(self, ns, n):
        nodes = ns.nodes()
        for id in nodes:
            ns.delete(id)

        nodeids = []
        ns.countdown(n * n, f"Testing {n}x{n} nodes ... %v seconds left")
        for r in range(n):
            for c in range(n):
                nodeid = ns.add("router", 50 + XGAP * c, 50 + YGAP * r, radio_range=RADIO_RANGE)
                nodeids.append(nodeid)

        secs = 0
        while secs < 1800:
            ns.go(PING_INTERVAL)
            secs += PING_INTERVAL
            self.produce_traffic(ns, nodeids)

    def produce_traffic(self, ns, nodeids):
        for p in ns.pings():
            print('ping', p)

        for i in range(PING_COUNT):
            n1 = random.choice(nodeids)
            while True:
                n2 = random.choice(nodeids)
                if n2 != n1:
                    break

            ns.ping(n1, n2, 'mleid', datasize=80)


if __name__ == '__main__':
    unittest.main()
