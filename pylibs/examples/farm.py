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
import math
import random

from otns.cli import OTNS

ns = OTNS(otns_args=['-log', 'info'])
ns.speed = 10

ns.web()
ns.visualization(broadcast_messages=False)

R = 7
RECEIVER_RADIO_RANGE = 300 * R
SENSOR_RADIO_RANGE = 100 * R
SENSOR_NUM = 10
FARM_RECT = [10 * R, 10 * R, 210 * R, 110 * R]

gateway = ns.add("router", FARM_RECT[0], FARM_RECT[1], radio_range=RECEIVER_RADIO_RANGE)
ns.add("router", FARM_RECT[0], FARM_RECT[3], radio_range=RECEIVER_RADIO_RANGE)
ns.add("router", FARM_RECT[2], FARM_RECT[1], radio_range=RECEIVER_RADIO_RANGE)
ns.add("router", FARM_RECT[2], FARM_RECT[3], radio_range=RECEIVER_RADIO_RANGE)
ns.add("router", (FARM_RECT[0] + FARM_RECT[2]) // 2, FARM_RECT[1], radio_range=RECEIVER_RADIO_RANGE)
ns.add("router", (FARM_RECT[0] + FARM_RECT[2]) // 2, FARM_RECT[3], radio_range=RECEIVER_RADIO_RANGE)

sensor_pos = {}
sensor_move_dir = {}

for i in range(SENSOR_NUM):
    rx = random.randint(FARM_RECT[0], FARM_RECT[2])
    ry = random.randint(FARM_RECT[1], FARM_RECT[3])
    sid = ns.add("sed", rx, ry, radio_range=SENSOR_RADIO_RANGE)
    sensor_pos[sid] = (rx, ry)
    sensor_move_dir[sid] = random.uniform(0, math.pi * 2)


def blocked(sid, x, y):
    if not (FARM_RECT[0] + 20 < x < FARM_RECT[2] - 20) or not (FARM_RECT[1] + 20 < y < FARM_RECT[3] - 20):
        return True

    for oid, (ox, oy) in sensor_pos.items():
        if oid == sid:
            continue

        dist2 = (x - ox) ** 2 + (y - oy) ** 2
        if dist2 <= 1600:
            return True

    return False


time_accum = 0
while True:
    dt = 1
    ns.go(dt)
    time_accum += dt

    for sid, (sx, sy) in sensor_pos.items():

        for i in range(10):
            mdist = random.uniform(0, 2 * R * dt)

            sx = int(sx + mdist * math.cos(sensor_move_dir[sid]))
            sy = int(sy + mdist * math.sin(sensor_move_dir[sid]))

            if blocked(sid, sx, sy):
                sensor_move_dir[sid] += random.uniform(0, math.pi * 2)
                continue

            sx = min(max(sx, FARM_RECT[0]), FARM_RECT[2])
            sy = min(max(sy, FARM_RECT[1]), FARM_RECT[3])
            ns.move(sid, sx, sy)

            sensor_pos[sid] = (sx, sy)
            break

    if time_accum >= 10:
        for sid in sensor_pos:
            ns.ping(sid, gateway)
        time_accum -= 10