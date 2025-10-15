package constant

const (
	ParticipantAddedSubject = "You’ve Been Invited to a Split Bill"

	ParticipantAddedTemplate = `
		<!DOCTYPE html>

		<html>

			<head>
				<title>You've Been Invited</title>
				<style>
					body {
						font-family: Arial, sans-serif;
						background-color: #f9f9f9;
						color: #333;
						margin: 0;
						padding: 0;
					}
					.email-container {
						max-width: 600px;
						margin: 30px auto;
						background: #fff;
						border: 1px solid #ddd;
						border-radius: 8px;
						padding: 24px;
					}
					h2 {
						color: #2b6cb0;
					}
					.button {
						display: inline-block;
						background-color: #2b6cb0;
						color: #fff;
						padding: 10px 18px;
						border-radius: 4px;
						text-decoration: none;
					}
					.button:hover {
						background-color: #1a4f8b;
					}
				</style>
			</head>

			<body>
				<div class="email-container">
					<h2>You’ve Been Invited to a Split Bill!</h2>
					<p>Hi {{.name}},</p>

					<p>You have been invited to join a split bill for <strong>{{.bill_title}}</strong>.</p>

					<p>Click the button below to view the bill details and confirm your participation:</p>

					<p>
						<a href="{{.action_url}}" class="button">View Split Bill</a>
					</p>

					<p>If you didn’t expect this invitation, you can safely ignore this email.</p>

					<p>Best regards,<br>SplitMyBill Team</p>
				</div>
			</body>

		</html>
    `
)
