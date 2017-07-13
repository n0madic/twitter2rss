<?php

/**
 * @file
 * A single location to store configuration.
 */

define('CONSUMER_KEY', getenv('CONSUMER_KEY') ? getenv('CONSUMER_KEY') : "");
define('CONSUMER_SECRET', getenv('CONSUMER_SECRET') ? getenv('CONSUMER_SECRET') : "");
define('USER_TOKEN', getenv('USER_TOKEN') ? getenv('USER_TOKEN') : "");
define('USER_SECRET', getenv('USER_SECRET') ? getenv('USER_SECRET') : "");
?>