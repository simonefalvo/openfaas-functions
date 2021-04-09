import os
import requests
import sys


def handle(req):
    """handle a request to the function
    Args:
        req (str): request body
    """

    # uses a default of "gateway" for when "gateway_hostname" is not set
    gateway_hostname = os.getenv("gateway_hostname", "gateway")
    result = req
    for i in (1, 2):

        resp = requests.get("http://" + gateway_hostname + ":8080/function/appender" + str(i), data=result)

        if resp.status_code != 200:
            sys.exit("Error with appender%d, expected: %d, got: %d\n" % (i, 200, resp.status_code))

        result = resp.text

    return result
