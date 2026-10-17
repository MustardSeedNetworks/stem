#!/bin/sh
set -e

# Offer automatic updates at install (ADR-0012). The question is asked once
# and defaults to off; a noninteractive install takes that default, and an
# unattended one can preseed the answer with debconf-set-selections.
. /usr/share/debconf/confmodule

db_input high stem/automatic-updates || true
db_go || true
