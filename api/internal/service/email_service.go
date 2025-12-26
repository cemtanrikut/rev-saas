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
	logoURL      string // Public URL for logo image
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
		logoURL:      "https://i.imgur.com/x8SISLw.png", // Revalyze logo hosted on Imgur
	}
}

// SetLogoURL sets the public URL for the logo image
func (s *EmailService) SetLogoURL(url string) {
	s.logoURL = url
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

	if err = c.Hello("localhost"); err != nil {
		return fmt.Errorf("EHLO failed: %w", err)
	}

	tlsconfig := &tls.Config{
		ServerName: s.config.Host,
	}
	if err = c.StartTLS(tlsconfig); err != nil {
		return fmt.Errorf("STARTTLS failed: %w", err)
	}

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

// getLogoHTML returns the logo HTML - either image or text fallback
func (s *EmailService) getLogoHTML() string {
	if s.logoURL != "" {
		return fmt.Sprintf(`<img src="%s" alt="Revalyze" width="140" height="auto" style="display: block; margin: 0 auto;" />`, s.logoURL)
	}
	// Text-based logo fallback with gradient background
	return `<table role="presentation" cellspacing="0" cellpadding="0" border="0" style="margin: 0 auto;">
              <tr>
                <td style="background: linear-gradient(135deg, #8b5cf6 0%, #d946ef 100%); padding: 12px 28px; border-radius: 12px;">
                  <span style="font-size: 26px; font-weight: 700; color: #ffffff; letter-spacing: -0.5px;">Revalyze</span>
                </td>
              </tr>
            </table>`
}

// SendVerificationEmail sends a verification email with a magic link
func (s *EmailService) SendVerificationEmail(ctx context.Context, toEmail, token string) error {
	verifyURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.appPublicURL, token)
	logoHTML := s.getLogoHTML()

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
              %s
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
                          <span style="font-size: 48px;">&#9993;&#65039;</span>
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
                            &#9200; This link expires in <strong>30 minutes</strong>
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
                &copy; 2025 Revalyze B.V. &bull; Utrecht, Netherlands
              </p>
            </td>
          </tr>
          
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, logoHTML, verifyURL, verifyURL)

	log.Printf("[email] Sending verification email to %s", toEmail)
	return s.sendEmail(ctx, toEmail, subject, htmlBody)
}

// SendWelcomeEmail sends a welcome email after verification
func (s *EmailService) SendWelcomeEmail(ctx context.Context, toEmail string) error {
	dashboardURL := fmt.Sprintf("%s/app/overview", s.appPublicURL)
	logoHTML := s.getLogoHTML()

	subject := "Welcome to Revalyze! 🎉"

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
        <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width: 600px; margin: 0 auto;">
          
          <!-- Celebration Header -->
          <tr>
            <td style="text-align: center; padding-bottom: 16px;">
              <span style="font-size: 40px;">&#127881; &#127882; &#127881;</span>
            </td>
          </tr>
          
          <!-- Logo -->
          <tr>
            <td style="text-align: center; padding-bottom: 32px;">
              %s
            </td>
          </tr>
          
          <!-- Main Card -->
          <tr>
            <td>
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 16px; border: 1px solid #334155;">
                <tr>
                  <td style="padding: 40px 32px;">
                    
                    <!-- Success Badge -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="text-align: center; padding-bottom: 20px;">
                          <span style="display: inline-block; background: linear-gradient(135deg, rgba(16, 185, 129, 0.2) 0%%, rgba(6, 182, 212, 0.2) 100%%); border: 1px solid rgba(16, 185, 129, 0.3); border-radius: 50px; padding: 8px 20px; font-size: 14px; color: #10b981; font-weight: 600;">
                            &#10003; Email Verified
                          </span>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Title -->
                    <h1 style="margin: 0 0 8px 0; font-size: 32px; font-weight: 700; color: #f8fafc; text-align: center; letter-spacing: -0.5px;">
                      Welcome aboard!
                    </h1>
                    
                    <!-- Subtitle -->
                    <p style="margin: 0 0 32px 0; font-size: 18px; line-height: 1.6; color: #94a3b8; text-align: center;">
                      You're all set to optimize your SaaS pricing strategy
                    </p>
                    
                    <!-- CTA Button -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="text-align: center; padding-bottom: 32px;">
                          <a href="%s" style="display: inline-block; padding: 18px 56px; background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); color: #ffffff; text-decoration: none; font-weight: 600; font-size: 16px; border-radius: 12px; box-shadow: 0 8px 24px -8px rgba(139, 92, 246, 0.5);">
                            Go to Dashboard &#8594;
                          </a>
                        </td>
                      </tr>
                    </table>
                    
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Section Title: What You Can Do -->
          <tr>
            <td style="padding: 32px 0 16px 0; text-align: center;">
              <h2 style="margin: 0; font-size: 20px; font-weight: 600; color: #f8fafc;">
                &#10024; What You Can Do
              </h2>
            </td>
          </tr>
          
          <!-- Feature Cards - 3 Separate Boxes -->
          <tr>
            <td>
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                
                <!-- Card 1: Discover Competitors -->
                <tr>
                  <td style="padding-bottom: 12px;">
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background: linear-gradient(135deg, #1e293b 0%%, #0f172a 100%%); border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 20px 24px;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                            <tr>
                              <td style="width: 48px; vertical-align: middle;">
                                <div style="width: 44px; height: 44px; background: linear-gradient(135deg, rgba(139, 92, 246, 0.2) 0%%, rgba(217, 70, 239, 0.2) 100%%); border-radius: 10px; text-align: center; line-height: 44px;">
                                  <span style="font-size: 22px;">&#128269;</span>
                                </div>
                              </td>
                              <td style="padding-left: 16px; vertical-align: middle;">
                                <p style="margin: 0 0 4px 0; font-size: 16px; font-weight: 600; color: #f8fafc;">Discover Competitors</p>
                                <p style="margin: 0; font-size: 14px; color: #94a3b8;">AI finds and tracks your competitors automatically</p>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                
                <!-- Card 2: Smart Insights -->
                <tr>
                  <td style="padding-bottom: 12px;">
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background: linear-gradient(135deg, #1e293b 0%%, #0f172a 100%%); border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 20px 24px;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                            <tr>
                              <td style="width: 48px; vertical-align: middle;">
                                <div style="width: 44px; height: 44px; background: linear-gradient(135deg, rgba(251, 191, 36, 0.2) 0%%, rgba(245, 158, 11, 0.2) 100%%); border-radius: 10px; text-align: center; line-height: 44px;">
                                  <span style="font-size: 22px;">&#128161;</span>
                                </div>
                              </td>
                              <td style="padding-left: 16px; vertical-align: middle;">
                                <p style="margin: 0 0 4px 0; font-size: 16px; font-weight: 600; color: #f8fafc;">Smart Insights</p>
                                <p style="margin: 0; font-size: 14px; color: #94a3b8;">Get data-driven pricing recommendations</p>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                
                <!-- Card 3: Simulations -->
                <tr>
                  <td style="padding-bottom: 12px;">
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background: linear-gradient(135deg, #1e293b 0%%, #0f172a 100%%); border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 20px 24px;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                            <tr>
                              <td style="width: 48px; vertical-align: middle;">
                                <div style="width: 44px; height: 44px; background: linear-gradient(135deg, rgba(16, 185, 129, 0.2) 0%%, rgba(6, 182, 212, 0.2) 100%%); border-radius: 10px; text-align: center; line-height: 44px;">
                                  <span style="font-size: 22px;">&#128200;</span>
                                </div>
                              </td>
                              <td style="padding-left: 16px; vertical-align: middle;">
                                <p style="margin: 0 0 4px 0; font-size: 16px; font-weight: 600; color: #f8fafc;">Simulations</p>
                                <p style="margin: 0; font-size: 14px; color: #94a3b8;">Test pricing changes before going live</p>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                
                <!-- Card 4: PDF Reports -->
                <tr>
                  <td>
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background: linear-gradient(135deg, #1e293b 0%%, #0f172a 100%%); border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 20px 24px;">
                          <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                            <tr>
                              <td style="width: 48px; vertical-align: middle;">
                                <div style="width: 44px; height: 44px; background: linear-gradient(135deg, rgba(239, 68, 68, 0.2) 0%%, rgba(244, 63, 94, 0.2) 100%%); border-radius: 10px; text-align: center; line-height: 44px;">
                                  <span style="font-size: 22px;">&#128196;</span>
                                </div>
                              </td>
                              <td style="padding-left: 16px; vertical-align: middle;">
                                <p style="margin: 0 0 4px 0; font-size: 16px; font-weight: 600; color: #f8fafc;">PDF Reports</p>
                                <p style="margin: 0; font-size: 14px; color: #94a3b8;">Export professional analysis reports</p>
                              </td>
                            </tr>
                          </table>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
                
              </table>
            </td>
          </tr>
          
          <!-- Section Title: Quick Start -->
          <tr>
            <td style="padding: 32px 0 16px 0; text-align: center;">
              <h2 style="margin: 0; font-size: 20px; font-weight: 600; color: #f8fafc;">
                &#128640; Quick Start Guide
              </h2>
            </td>
          </tr>
          
          <!-- Steps Card -->
          <tr>
            <td>
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 16px; border: 1px solid #334155;">
                <tr>
                  <td style="padding: 28px 32px;">
                    
                    <!-- Step 1 -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="padding-bottom: 20px;">
                      <tr>
                        <td style="width: 36px; vertical-align: top;">
                          <div style="width: 28px; height: 28px; background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); border-radius: 50%%; text-align: center; line-height: 28px; font-size: 14px; font-weight: 700; color: #fff;">1</div>
                        </td>
                        <td style="padding-left: 12px; vertical-align: top;">
                          <p style="margin: 0 0 2px 0; font-size: 15px; font-weight: 600; color: #f8fafc;">Add Your Pricing Plans</p>
                          <p style="margin: 0; font-size: 13px; color: #94a3b8;">Import from your website or enter manually</p>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Step 2 -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="padding-bottom: 20px;">
                      <tr>
                        <td style="width: 36px; vertical-align: top;">
                          <div style="width: 28px; height: 28px; background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); border-radius: 50%%; text-align: center; line-height: 28px; font-size: 14px; font-weight: 700; color: #fff;">2</div>
                        </td>
                        <td style="padding-left: 12px; vertical-align: top;">
                          <p style="margin: 0 0 2px 0; font-size: 15px; font-weight: 600; color: #f8fafc;">Discover Competitors</p>
                          <p style="margin: 0; font-size: 13px; color: #94a3b8;">Let AI find and analyze your market</p>
                        </td>
                      </tr>
                    </table>
                    
                    <!-- Step 3 -->
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                      <tr>
                        <td style="width: 36px; vertical-align: top;">
                          <div style="width: 28px; height: 28px; background: linear-gradient(135deg, #8b5cf6 0%%, #d946ef 100%%); border-radius: 50%%; text-align: center; line-height: 28px; font-size: 14px; font-weight: 700; color: #fff;">3</div>
                        </td>
                        <td style="padding-left: 12px; vertical-align: top;">
                          <p style="margin: 0 0 2px 0; font-size: 15px; font-weight: 600; color: #f8fafc;">Get AI Analysis</p>
                          <p style="margin: 0; font-size: 13px; color: #94a3b8;">Receive actionable pricing insights</p>
                        </td>
                      </tr>
                    </table>
                    
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Stats Row -->
          <tr>
            <td style="padding-top: 24px;">
              <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
                <tr>
                  <td style="width: 33.33%%; padding-right: 8px;">
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 16px; text-align: center;">
                          <p style="margin: 0 0 4px 0; font-size: 24px; font-weight: 700; color: #8b5cf6;">10+</p>
                          <p style="margin: 0; font-size: 12px; color: #94a3b8;">Free Credits</p>
                        </td>
                      </tr>
                    </table>
                  </td>
                  <td style="width: 33.33%%; padding: 0 4px;">
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 16px; text-align: center;">
                          <p style="margin: 0 0 4px 0; font-size: 24px; font-weight: 700; color: #10b981;">5min</p>
                          <p style="margin: 0; font-size: 12px; color: #94a3b8;">Setup Time</p>
                        </td>
                      </tr>
                    </table>
                  </td>
                  <td style="width: 33.33%%; padding-left: 8px;">
                    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #1e293b; border-radius: 12px; border: 1px solid #334155;">
                      <tr>
                        <td style="padding: 16px; text-align: center;">
                          <p style="margin: 0 0 4px 0; font-size: 24px; font-weight: 700; color: #fbbf24;">24/7</p>
                          <p style="margin: 0; font-size: 12px; color: #94a3b8;">AI Available</p>
                        </td>
                      </tr>
                    </table>
                  </td>
                </tr>
              </table>
            </td>
          </tr>
          
          <!-- Bottom Celebration -->
          <tr>
            <td style="text-align: center; padding-top: 32px;">
              <span style="font-size: 28px;">&#127878; &#127880; &#127873; &#127880; &#127878;</span>
            </td>
          </tr>
          
          <!-- Footer -->
          <tr>
            <td style="padding-top: 32px; text-align: center;">
              <p style="margin: 0 0 8px 0; font-size: 13px; color: #64748b; line-height: 1.5;">
                Questions? Just reply to this email — we're here to help!
              </p>
              <p style="margin: 16px 0 0 0; font-size: 12px; color: #475569;">
                &copy; 2025 Revalyze B.V. &bull; Utrecht, Netherlands
              </p>
            </td>
          </tr>
          
        </table>
      </td>
    </tr>
  </table>
</body>
</html>`, logoHTML, dashboardURL)

	log.Printf("[email] Sending welcome email to %s", toEmail)
	return s.sendEmail(ctx, toEmail, subject, htmlBody)
}
