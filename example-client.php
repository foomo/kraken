<?php

/*
 * This file is part of the foomo Opensource Framework.
 *
 * The foomo Opensource Framework is free software: you can redistribute it
 * and/or modify it under the terms of the GNU Lesser General Public License as
 * published  by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * The foomo Opensource Framework is distributed in the hope that it will
 * be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
 * of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License along with
 * the foomo Opensource Framework. If not, see <http://www.gnu.org/licenses/>.
 */


/**
 * @link www.foomo.org
 * @license www.gnu.org/licenses/lgpl.txt
 */
class Kraken
{
	private $server = '';
	public function __construct($server)
	{
		$this->server = $server;
	}
	public function createTentacle($name, $bandwidth, $retry)
	{
		return $this->callServer(
			'PUT',
			'/tentacle/' . urlencode($name), [
				'bandwidth' => $bandwidth,
				'retry' => $retry
			]
		);
	}
	public function deleteTentacle($name)
	{
		$this->callServer(
			'DELETE',
			'/tentacle/' . urlencode($name)
		);
	}

	public function reset()
	{
		$status = $this->getStatus();
		foreach($status->tentacles as $tentacleStatus) {
			$this->deleteTentacle($tentacleStatus->name);
		}
	}
	public function addPrey($tentacle, $id, $url, $priority, $body = null, $method = 'GET', array $tags = [])
	{
		$callData = [
			'url'      => $url,
			'priority' => $priority,
			'method'   => $method,
			'tags'     => $tags
		];
		if(!is_null($body)) {
			$callData['body'] = $body;
		}
		return $this->callServer(
			'PUT',
			'/tentacle/' . urlencode($tentacle) . '/' . urlencode($id),
			$callData
		);
	}

	public function getStatus()
	{
		return $this->callServer('GET', '/status', null);
	}

	public function getStatusForTentacle($name)
	{
		return $this->callServer('GET', '/tentacle/' . urlencode($name), null);
	}


	public function callServer($method, $path, $request = null)
	{
		$opts = ['http' =>
			[
				'method'  => $method,
				'header'  => 'Content-type: application/json'
			]
		];
		if(!is_null($request)) {
			$opts['http']['content'] = json_encode($request);
		}
		$context = stream_context_create($opts);
		return json_decode(file_get_contents($this->server . $path, false, $context));
	}
}


$kraken = new Kraken('http://127.0.0.1:8888');
// create a tentacle
var_dump($kraken->createTentacle($tentacleName = 'test', 4, 2));

// add some prey
for($i = 0;$i<10;$i++) {
	$kraken->addPrey($tentacleName, 'test-' . $i, 'https://www.google.com/robots.txt', 100);
}

// kill it all
while(true) {
	$status = $kraken->getStatusForTentacle($tentacleName);
	foreach($status->prey as $p) {
		if($p->status != "done") {
			echo "waiting for " . $p->id . PHP_EOL;
			continue 2;
		} else {
			echo "done with " . $p->id . PHP_EOL;
		}
	}
	echo "done" . PHP_EOL;
	var_dump($kraken->getStatus());
	$kraken->deleteTentacle($tentacleName);
	exit;
}
