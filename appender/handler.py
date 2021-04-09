import os
import logging


def handle(req):
    """handle a request to the function
    Args:
        req (str): request body
    """
    logging.info("Received message: " + req)
    msg = os.getenv("message", " - handled by appender") 

    return req.rstrip() + msg
