package worker

var emailStyle = `
		body {
			font-family: Arial, sans-serif;
			background-color: #f4f4f4;
			margin: 0;
			padding: 0;
		}
		.container {
			background-color: #ffffff;
			margin: 50px auto;
			padding: 20px;
			border-radius: 8px;
			box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
			max-width: 600px;
		}
		h1 {
			color: #333333;
		}
		p {
			color: #555555;
			line-height: 1.6;
		}
		.button {
			display: inline-block;
			padding: 10px 20px;
			margin-top: 20px;
			background-color: #28a745;
			color: #ffffff;
			text-decoration: none;
			border-radius: 5px;
		}
		.footer {
			margin-top: 30px;
			font-size: 12px;
			color: #888888;
		}
	`

func EmailTemplate(userName, documentName, expirationDate string) string {
	return `
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Document Expiration Reminder</title>
			<style>
				` + emailStyle + `
			</style>
		</head>
		<body>
			<div class="container">
				<h1>Reminder: Your Document is Expiring Soon</h1>
				<p>Hi ` + userName + `,</p>
				<p>This is a friendly reminder that your document "<strong>` + documentName + `</strong>" is set to expire on <strong>` + expirationDate + `</strong>.</p>
				<p>Please take the necessary actions to renew or update your document before the expiration date to avoid any disruptions.</p>
				<a href="#" class="button">Manage Your Documents</a>
				<p class="footer">If you have any questions, feel free to contact our support team.</p>
			</div>
		</body>
		</html>
	`
}

func SMSMessage(documentName, expirationDate string) string {
	return "Reminder: Your document '" + documentName + "' is expiring on " + expirationDate + ". Please take action to renew it."
}
