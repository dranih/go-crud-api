package middleware

import (
	"log"
	"net/http"
)

type BasicAuthMiddleware struct {
	GenericMiddleware
}

func (bam *BasicAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

/*
class BasicAuthMiddleware extends Middleware
{
	private function hasCorrectPassword(string $username, string $password, array &$passwords): bool
	{
		$hash = isset($passwords[$username]) ? $passwords[$username] : false;
		if ($hash && password_verify($password, $hash)) {
			if (password_needs_rehash($hash, PASSWORD_DEFAULT)) {
				$passwords[$username] = password_hash($password, PASSWORD_DEFAULT);
			}
			return true;
		}
		return false;
	}

	private function getValidUsername(string $username, string $password, string $passwordFile): string
	{
		$passwords = $this->readPasswords($passwordFile);
		$valid = $this->hasCorrectPassword($username, $password, $passwords);
		$this->writePasswords($passwordFile, $passwords);
		return $valid ? $username : '';
	}

	private function readPasswords(string $passwordFile): array
	{
		$passwords = [];
		$passwordLines = file($passwordFile);
		foreach ($passwordLines as $passwordLine) {
			if (strpos($passwordLine, ':') !== false) {
				list($username, $hash) = explode(':', trim($passwordLine), 2);
				if (strlen($hash) > 0 && $hash[0] != '$') {
					$hash = password_hash($hash, PASSWORD_DEFAULT);
				}
				$passwords[$username] = $hash;
			}
		}
		return $passwords;
	}

	private function writePasswords(string $passwordFile, array $passwords): bool
	{
		$success = false;
		$passwordFileContents = '';
		foreach ($passwords as $username => $hash) {
			$passwordFileContents .= "$username:$hash\n";
		}
		if (file_get_contents($passwordFile) != $passwordFileContents) {
			$success = file_put_contents($passwordFile, $passwordFileContents) !== false;
		}
		return $success;
	}

	private function getAuthorizationCredentials(ServerRequestInterface $request): string
	{
		$serverParams = $request->getServerParams();
		if (isset($serverParams['PHP_AUTH_USER'])) {
			return $serverParams['PHP_AUTH_USER'] . ':' . $serverParams['PHP_AUTH_PW'];
		}
		$header = RequestUtils::getHeader($request, 'Authorization');
		$parts = explode(' ', trim($header), 2);
		if (count($parts) != 2) {
			return '';
		}
		if ($parts[0] != 'Basic') {
			return '';
		}
		return base64_decode(strtr($parts[1], '-_', '+/'));
	}

	public function process(ServerRequestInterface $request, RequestHandlerInterface $next): ResponseInterface
	{
		if (session_status() == PHP_SESSION_NONE) {
			if (!headers_sent()) {
				$sessionName = $this->getProperty('sessionName', '');
				if ($sessionName) {
					session_name($sessionName);
				}
				session_start();
			}
		}
		$credentials = $this->getAuthorizationCredentials($request);
		if ($credentials) {
			list($username, $password) = array('', '');
			if (strpos($credentials, ':') !== false) {
				list($username, $password) = explode(':', $credentials, 2);
			}
			$passwordFile = $this->getProperty('passwordFile', '.htpasswd');
			$validUser = $this->getValidUsername($username, $password, $passwordFile);
			$_SESSION['username'] = $validUser;
			if (!$validUser) {
				return $this->responder->error(ErrorCode::AUTHENTICATION_FAILED, $username);
			}
			if (!headers_sent()) {
				session_regenerate_id();
			}
		}
		if (!isset($_SESSION['username']) || !$_SESSION['username']) {
			$authenticationMode = $this->getProperty('mode', 'required');
			if ($authenticationMode == 'required') {
				$response = $this->responder->error(ErrorCode::AUTHENTICATION_REQUIRED, '');
				$realm = $this->getProperty('realm', 'Username and password required');
				$response = $response->withHeader('WWW-Authenticate', "Basic realm=\"$realm\"");
				return $response;
			}
		}
		return $next->handle($request);
	}
}*/
