package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/smtp"
	"strings"
)

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	From     string
}

// EmailService handles sending emails via SMTP
type EmailService struct {
	config       SMTPConfig
	appPublicURL string
}

// NewEmailService creates a new EmailService
func NewEmailService(host, port, user, password, from, appPublicURL string) *EmailService {
	return &EmailService{
		config: SMTPConfig{
			Host:     host,
			Port:     port,
			User:     user,
			Password: password,
			From:     from,
		},
		appPublicURL: appPublicURL,
	}
}

// sendEmail sends an email using the configured SMTP server.
// Supports both STARTTLS (port 587) and direct TLS (port 465)
func (s *EmailService) sendEmail(ctx context.Context, to, subject, htmlBody string) error {
	if s.config.Host == "" || s.config.Port == "" {
		return fmt.Errorf("SMTP not configured")
	}

	addr := fmt.Sprintf("%s:%s", s.config.Host, s.config.Port)

	// Build the email message with headers
	// Extract email address from "Name <email>" format if present
	fromEmail := s.config.From
	if idx := strings.Index(fromEmail, "<"); idx != -1 {
		fromEmail = strings.TrimSuffix(fromEmail[idx+1:], ">")
	}

	msg := fmt.Sprintf("From: %s\r\n", s.config.From)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += htmlBody

	// Set up authentication
	auth := smtp.PlainAuth("", s.config.User, s.config.Password, s.config.Host)

	// Use different connection method based on port
	if s.config.Port == "465" {
		// Direct TLS connection (SSL)
		return s.sendWithDirectTLS(addr, auth, fromEmail, to, []byte(msg))
	}

	// STARTTLS connection (port 587 or 25)
	return s.sendWithSTARTTLS(addr, auth, fromEmail, to, []byte(msg))
}

// sendWithSTARTTLS connects using STARTTLS (for port 587)
func (s *EmailService) sendWithSTARTTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	// Connect to the server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	c, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer c.Close()

	// Say EHLO
	if err = c.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO failed: %w", err)
	}

	// Start TLS
	tlsconfig := &tls.Config{
		ServerName: s.config.Host,
	}
	if err = c.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

	// Authenticate
	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	// Set sender and recipient
	if err = c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	// Send the message body
	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return c.Quit()
}

// sendWithDirectTLS connects using direct TLS (for port 465)
func (s *EmailService) sendWithDirectTLS(addr string, auth smtp.Auth, from, to string, msg []byte) error {
	tlsconfig := &tls.Config{
		ServerName: s.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsconfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}

	c, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer c.Close()

	if err = c.Auth(auth); err != nil {
		return fmt.Errorf("SMTP auth failed: %w", err)
	}

	if err = c.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}

	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TO failed: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return c.Quit()
}

// SendVerificationEmail sends a verification email with a magic link
func (s *EmailService) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.appPublicURL, token)

	subject := "Verify your email - Revalyze"

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Verify Your Email</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0f172a; -webkit-font-smoothing: antialiased;">
  <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #0f172a;">
    <tr>
      <td style="padding: 48px 24px;">
        <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width: 520px; margin: 0 auto;">
          
          <!-- Logo -->
          <tr>
            <td style="text-align: center; padding-bottom: 32px;">
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 0 auto;">
                <tr>
                  <td style="background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); padding: 12px 24px; border-radius: 12px;">
                    <span style="font-size: 24px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px;">Revalyze</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Main Card -->
          <tr>
            <td>
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 16px; border: 1px solid #334155;">
                <tr>
                  <td style="padding: 40px 32px;">
                    
                    <!-- Icon -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="text-align: center; padding-bottom: 24px;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 0 auto;">
                            <tr>
                              <td style="width: 64px; height: 64px; background: linear-gradient(135deg, rgba(139, 92, 246, 0.2) 0%%, rgba(217, 70, 239, 0.2) 100%%); border-radius: 16px; text-align: center; vertical-align: middle; border: 1px solid rgba(139, 92, 246, 0.3);">
                                <span style="font-size: 28px; line-height: 64px;">&#9993;</span>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Title -->
                    <h1 style="margin: 0 0 12px 0; font-size: 28px; font-weight: 700; color: #f8fafc; text-align: center; letter-spacing: -0.5px;">
                      Verify Your Email
                    </h1>
                    
                    <!-- Description -->
                    <p style="margin: 0 0 32px 0; font-size: 16px; line-height: 1.7; color: #94a3b8; text-align: center;">
                      Thanks for signing up! Click the button below to verify your email address and start optimizing your pricing strategy.
                    </p>
                    
                    <!-- CTA Button -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="text-align: center; padding-bottom: 32px;">
                          <a href="%s" style="display: inline-block; padding: 16px 40px; background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); color: #ffffff; text-decoration: none; font-weight: 600; font-size: 16px; border-radius: 12px;">
                            Verify Email Address
                          </a>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Expiry Note -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="background-color: rgba(251, 191, 36, 0.1); border: 1px solid rgba(251, 191, 36, 0.2); border-radius: 12px; padding: 16px 20px; text-align: center;">
                          <p style="margin: 0; font-size: 14px; color: #fbbf24;">
                            This link expires in <strong>30 minutes</strong>
                          </p>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Divider -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="padding: 28px 0;">
                          <hr style="border: none; border-top: 1px solid #334155; margin: 0;" />
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Alternative Link -->
                    <p style="margin: 0 0 8px 0; font-size: 13px; color: #64748b; text-align: center;">
                      If the button doesn't work, copy and paste this link:
                    </p>
                    <p style="margin: 0; font-size: 12px; color: #8b5cf6; word-break: break-all; text-align: center; background-color: #0f172a; padding: 12px 16px; border-radius: 8px; font-family: monospace;">
                      %s
                    </p>
                    
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Footer -->
          <tr>
            <td style="padding-top: 32px; text-align: center;">
              <p style="margin: 0 0 8px 0; font-size: 13px; color: #64748b; line-height: 1.5;">
                Didn't sign up for Revalyze? You can safely ignore this email.
              </p>
              <p style="margin: 16px 0 0 0; font-size: 12px; color: #475569;">
                &copy; 2025 Revalyze B.V. &bull; Amsterdam, Netherlands
              </p>
            </td>
          </tr>
          
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, verifyURL, verifyURL)

	log.Printf("[email] Sending verification email to %s", toEmail)
	return s.sendEmail(ctx, toEmail, subject, htmlBody)
}

// SendWelcomeEmail sends a welcome email after verification
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail string) error {
	dashboardURL := fmt.Sprintf("%s/app/overview", s.appPublicURL)

	subject := "Welcome to Revalyze"

	htmlBody := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Welcome to Revalyze</title>
</head>
<body style="margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif; background-color: #0f172a; -webkit-font-smoothing: antialiased;">
  <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #0f172a;">
    <tr>
      <td style="padding: 48px 24px;">
        <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width: 520px; margin: 0 auto;">
          
          <!-- Logo -->
          <tr>
            <td style="text-align: center; padding-bottom: 32px;">
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 0 auto;">
                <tr>
                  <td style="background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); padding: 12px 24px; border-radius: 12px;">
                    <span style="font-size: 24px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px;">Revalyze</span>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Main Card -->
          <tr>
            <td>
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 16px; border: 1px solid #334155;">
                <tr>
                  <td style="padding: 40px 32px;">
                    
                    <!-- Success Icon -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="text-align: center; padding-bottom: 24px;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 0 auto;">
                            <tr>
                              <td style="width: 64px; height: 64px; background: linear-gradient(135deg, rgba(16, 185, 129, 0.2) 0%%, rgba(6, 182, 212, 0.2) 100%%); border-radius: 16px; text-align: center; vertical-align: middle; border: 1px solid rgba(16, 185, 129, 0.3);">
                                <span style="font-size: 28px; line-height: 64px; color: #10b981;">&#10003;</span>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Title -->
                    <h1 style="margin: 0 0 12px 0; font-size: 28px; font-weight: 700; color: #f8fafc; text-align: center; letter-spacing: -0.5px;">
                      Welcome to Revalyze!
                    </h1>
                    
                    <!-- Description -->
                    <p style="margin: 0 0 32px 0; font-size: 16px; line-height: 1.7; color: #94a3b8; text-align: center;">
                      Your email is verified and your account is ready. Start optimizing your SaaS pricing strategy with AI-powered insights.
                    </p>
                    
                    <!-- Features -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 32px;">
                      <tr>
                        <td style="padding: 16px; background-color: #0f172a; border-radius: 12px; border: 1px solid #334155;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                            <tr>
                              <td style="padding-bottom: 12px;">
                                <table role="presentation" cellspacing="0" cellpadding="0" border="0">
                                  <tr>
                                    <td style="width: 24px; vertical-align: top; padding-right: 12px;">
                                      <span style="color: #10b981; font-size: 14px;">&#10003;</span>
                                    </td>
                                    <td style="font-size: 14px; color: #cbd5e1; line-height: 1.5;">
                                      <strong style="color: #f8fafc;">Competitive Analysis</strong> - Track competitor pricing
                                    </td>
                                  </tr>
                                </table>
                              </td>
                            </tr>
                            <tr>
                              <td style="padding-bottom: 12px;">
                                <table role="presentation" cellspacing="0" cellpadding="0" border="0">
                                  <tr>
                                    <td style="width: 24px; vertical-align: top; padding-right: 12px;">
                                      <span style="color: #10b981; font-size: 14px;">&#10003;</span>
                                    </td>
                                    <td style="font-size: 14px; color: #cbd5e1; line-height: 1.5;">
                                      <strong style="color: #f8fafc;">AI Insights</strong> - Smart pricing recommendations
                                    </td>
                                  </tr>
                                </table>
                              </td>
                            </tr>
                            <tr>
                              <td>
                                <table role="presentation" cellspacing="0" cellpadding="0" border="0">
                                  <tr>
                                    <td style="width: 24px; vertical-align: top; padding-right: 12px;">
                                      <span style="color: #10b981; font-size: 14px;">&#10003;</span>
                                    </td>
                                    <td style="font-size: 14px; color: #cbd5e1; line-height: 1.5;">
                                      <strong style="color: #f8fafc;">Simulations</strong> - Test scenarios before changes
                                    </td>
                                  </tr>
                                </table>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- CTA Button -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="text-align: center;">
                          <a href="%s" style="display: inline-block; padding: 16px 40px; background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); color: #ffffff; text-decoration: none; font-weight: 600; font-size: 16px; border-radius: 12px;">
                            Go to Dashboard
                          </a>
                        </td>
                      </tr>
                    </table>
                    
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Footer -->
          <tr>
            <td style="padding-top: 32px; text-align: center;">
              <p style="margin: 0 0 8px 0; font-size: 13px; color: #64748b; line-height: 1.5;">
                Need help? Reply to this email and we'll assist you.
              </p>
              <p style="margin: 16px 0 0 0; font-size: 12px; color: #475569;">
                &copy; 2025 Revalyze B.V. &bull; Amsterdam, Netherlands
              </p>
            </td>
          </tr>
          
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, dashboardURL)

	log.Printf("[email] Sending welcome email to %s", toEmail)
	return s.sendEmail(ctx, toEmail, subject, htmlBody)
}
