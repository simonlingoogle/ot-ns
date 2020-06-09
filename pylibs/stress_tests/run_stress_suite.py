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

import inspect
import logging
import sys
import os

from stess_report_bot import StressReportBot


def find_stress_test_classes(mod, suite_name: str):
    from BaseStressTest import BaseStressTest

    sts = []
    for name, member in inspect.getmembers(mod):
        if not isinstance(member, type):
            continue

        if issubclass(member, BaseStressTest) and member is not BaseStressTest:
            assert hasattr(member, 'SUITE') and isinstance(member.SUITE,
                                                           str), f'Please define `SUITE` for {member.__name__}'
            if member.SUITE == suite_name:
                sts.append(member)

    return sts


def main():
    logging.basicConfig(format='%(asctime)s - %(levelname)s - %(message)s', level=logging.DEBUG)

    script_path = sys.argv[0]
    script_dir = os.path.abspath(os.path.dirname(script_path))
    logging.info('script directory: %s' % script_dir)

    suite_names = sys.argv[1:]
    if not suite_names:
        logging.error("suite names not specified, nothing to do")
        exit(-1)

    logging.info("Running stress test suite: %s ...", ', '.join(suite_names))

    sys.path.insert(0, script_dir)

    for suite_name in suite_names:
        run_suite(script_dir, suite_name)


def run_suite(script_dir, suite_name: str):
    stress_tests = []
    for filename in os.listdir(script_dir):
        if not filename.endswith('.py'):
            continue

        modname = os.path.splitext(filename)[0]
        mod = __import__(modname)
        stress_test_classes = find_stress_test_classes(mod, suite_name)
        if not stress_test_classes:
            continue

        stress_tests.append((filename, stress_test_classes))

    results = []

    for filename, clses in sorted(stress_tests):
        for cls in clses:
            t = cls()
            logging.info("Running stress test: %s ...", t.name)
            t.run(report=False)
            results.append(t.result)

    # results.sort(key=lambda r: (0 if r.failed else 1, r.name))
    #
    # passed_count = sum(1 for result in results if not result.failed)
    # failed_count = sum(1 for result in results if result.failed)

    installid, owner, repo, issue_number = None, None, None, None
    if os.getenv('GITHUB_ACTIONS'):
        installid = int(os.environ['OT_STRESS_REPORT_INSTALL_ID'])
    else:
        installid = 8493459
        owner, repo, issue_number = 'simonlingoogle', 'ot-ns', 1

    bot = StressReportBot(installid=installid, owner=owner, repo=repo, issue_number=issue_number)
    bot.submit_suite_results(suite_name, results)

    # with StringIO() as report:
    #     report.write(
    #         f'''**[OTNS](https://github.com/openthread/ot-ns) Stress Tests Report Generated at {time.strftime(
    #             "%m/%d %H:%M:%S")}: {passed_count} passed, {failed_count} failed**\n''')
    #     for result in results:
    #         report.write(result.format())

    # ret = bot.comment_once(report.getvalue())


if __name__ == '__main__':
    main()