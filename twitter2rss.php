<?php
if (in_array($_SERVER["SERVER_ADDR"], ['127.0.0.1', 'dev.nomadic.name'])) {
	ini_set('display_errors', 'On');
	error_reporting(E_ALL);
} else {
	ini_set('display_errors', 'Off');
	error_reporting(E_ERROR);
}

require_once('tmhOAuth.php');
require_once('config.php');

mb_internal_encoding("UTF-8");

if (!empty($_REQUEST['name'])) {

	$screen_name = trim($_REQUEST['name']);
	$count = (!empty($_REQUEST['count'])) ? (($_REQUEST['count'] <= 200) ? $_REQUEST['count'] : '200') : '20';
	$exclude_replies = (!empty($_REQUEST['exclude_replies'])) ? 'true' : 'false';

	$tmhOAuth = new tmhOAuth(array(
		'consumer_key' => CONSUMER_KEY,
		'consumer_secret' => CONSUMER_SECRET,
		'token' => USER_TOKEN,
		'secret' => USER_SECRET,
	));

	$code = $tmhOAuth->request('GET', $tmhOAuth->url('1.1/statuses/user_timeline.json'),
									array('include_entities' => 'false',
										'include_rts' => 'true',
										'trim_user' => 'true',
										'screen_name' => $screen_name,
										'exclude_replies' => $exclude_replies,
										'count' => $count), true);
	if ($code == 200) {
		$responseData = json_decode($tmhOAuth->response['response'], true);
		if (isset($_REQUEST['response'])) {
			echo '<h1>[DEBUG] Response from twitter server:</h1>';
			header('Content-Type: text/html; charset=utf-8');
			var_dump($responseData);
			die();
		}
		header('Content-Type: application/atom+xml; charset=utf-8');
		echo '<?xml version="1.0" encoding="utf-8"?>' . PHP_EOL;
		echo '<feed xmlns="http://www.w3.org/2005/Atom">' . PHP_EOL;
		echo '<id>tag:twitter.com,' . date('Y-m-d') . ':' . $screen_name . '</id>' . PHP_EOL;
		echo '<title>Twitter feed @' . $screen_name . '</title>' . PHP_EOL;
		echo '<author><name>' . $screen_name . '</name></author>' . PHP_EOL;
		echo '<link type="application/atom+xml" href="' . (isset($_SERVER['HTTPS']) ? 'https' : 'http') . '://' . $_SERVER['HTTP_HOST'] . $_SERVER['REQUEST_URI'] . '" rel="self"/>' . PHP_EOL;
		echo '<link type="text/html" href="https://twitter.com/' . $screen_name . '" rel="alternate"/>' . PHP_EOL;
		$date = new DateTime($responseData[0]['created_at']);
		echo '<updated>' . $date->format("Y-m-d\TH:i:s\Z") . '</updated>' . PHP_EOL;
		foreach ($responseData as $tweet) {
			echo '<entry>' . PHP_EOL;
			$title = preg_split("/\r\n|\n|\r/", $tweet['text'], -1, PREG_SPLIT_NO_EMPTY);
			echo '<title>' . htmlspecialchars(preg_replace("/:$/", "$1", trim(preg_replace('/^(.*?)(?=http:\/\/t.co|([.?!]\s|$)).+/', '$1', $title[0])))) . '</title>' . PHP_EOL;
			echo '<author><name>' . $screen_name . '</name></author>' . PHP_EOL;
			echo '<id>tag:twitter.com,' . date('Y-m-d') . ':' . $screen_name . '/statuses/' . $tweet['id'] . '</id>' . PHP_EOL;
			echo '<updated>' . date('c', strtotime($tweet['created_at'])) . '</updated>' . PHP_EOL;
			echo '<link href="https://twitter.com/' . $screen_name . '/statuses/' . $tweet['id'] . '"/>' . PHP_EOL;
			$text = preg_replace('/(http:\/\/t\.co\/\w+)(?=\s|$)/', '<a href=$1>$1</a>', $tweet['text']);
			echo '<summary type="html"><![CDATA[' . $text;
			if (isset($tweet['extended_entities']['media'])) {
				 echo '<br />';
				foreach ($tweet['extended_entities']['media'] as $media) {
					echo '<img src="' . $media['media_url'] . '">';
					echo '<br />';
				}
			}
			echo ']]></summary>' . PHP_EOL;
			echo '</entry>' . PHP_EOL;
		}
		echo '</feed>' . PHP_EOL;
		die();
	} else {
		http_response_code(404);
		header('Content-Type: text/plain; charset=utf-8');
		die('ERROR get twitter\'s timeline for ' . $screen_name);
	}
}
?>
<!DOCTYPE html>
<html>
<head>
	<title>Twitter to RSS proxy</title>
	<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.2.0/css/bootstrap.min.css">
</head>
<body style="padding: 20px;">
	<div class="container">
		<div class="jumbotron vertical-center">
			<div class="container">
				<h1>Twitter to RSS
					<small>proxy</small>
				</h1>
				<p>Enter Twitter name and get full RSS feed!</p>
				<form id="tform" class="form-horizontal" role="form" action="<?php echo $_SERVER['REQUEST_URI']; ?>" method="GET">
					<div class="form-group">
						<div class="input-group input-group-lg">
							<span class="input-group-addon">@</span>
							<input type="text" id="name" name="name" class="form-control search-query" placeholder="Twitter name" required>
								<span class="input-group-btn">
									<input class="btn btn-primary" type="submit" value="Get RSS">
								</span>
						</div>
						<div class="panel panel-default" style="margin-top: 20px;">
							<div class="panel-body">
								<div class="form-group" style="margin-bottom: 0;">
									<label for="count" class="col-sm-3 control-label">Number of tweets (max 200):</label>
									<div class="col-sm-2">
										<input name="count" id="count" class="form-control" placeholder="20">
									</div>
									<input style="margin-top: 10px;" name="exclude_replies" type="checkbox"> Exclude Replies
								</div>
							</div>
						</div>
					</div>
				</form>
			</div>
		</div> <!-- jumbotron -->
		<footer class="navbar-fixed-bottom">
			<div style="text-align: center;"><p>&copy; Nomadic 2014</p></div>
		</footer>
	</div>
	<script>
		var Form = document.getElementById('tform');
		Form.onsubmit = function(event) {
			var count_input = document.getElementById('count');
			if (count_input.getAttribute('name') && !count_input.value) {
					count_input.setAttribute('name', '');
				}
			var screen_name = document.getElementById('name');
			if (screen_name.value != '') {
				Form.action = screen_name.value;
				screen_name.setAttribute('name', '');
			}
		};
	</script>
</body>
</html>
