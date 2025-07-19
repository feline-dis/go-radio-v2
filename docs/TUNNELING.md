# External Access and Tunneling

Go Radio v2 runs locally on your machine and is accessible at `http://localhost:8080` by default. To make your radio accessible from the internet, you'll need to set up a tunneling service.

## What is Tunneling?

Tunneling allows you to expose your local Go Radio instance to the internet by creating a secure connection between your local machine and a public endpoint. This enables others to access your radio from anywhere.

## Recommended Tunneling Services

### 1. ngrok (Recommended)

[ngrok](https://ngrok.com/) is a popular tunneling service that's easy to set up and use.

#### Installation

**macOS (via Homebrew):**
```bash
brew install ngrok/ngrok/ngrok
```

**Linux/Windows:**
Download from [ngrok.com/download](https://ngrok.com/download)

#### Setup

1. Sign up for a free account at [ngrok.com](https://ngrok.com/)
2. Get your auth token from the [ngrok dashboard](https://dashboard.ngrok.com/get-started/your-authtoken)
3. Configure ngrok with your auth token:
   ```bash
   ngrok config add-authtoken YOUR_AUTH_TOKEN
   ```

#### Usage

1. Start your Go Radio server:
   ```bash
   make run
   # or
   ./bin/go-radio-server
   ```

2. In a new terminal, start ngrok:
   ```bash
   ngrok http 8080
   ```

3. ngrok will display a public URL (e.g., `https://abc123.ngrok.io`) that forwards to your local radio

#### Custom Domain (Paid Plan)

If you have an ngrok paid plan, you can use a custom domain:
```bash
ngrok http --domain=your-custom-domain.ngrok.io 8080
```

### 2. Cloudflare Tunnel

[Cloudflare Tunnel](https://developers.cloudflare.com/cloudflare-one/connections/connect-apps/) is a free alternative that requires a Cloudflare account.

#### Installation

```bash
# macOS
brew install cloudflare/cloudflare/cloudflared

# Linux
wget -q https://github.com/cloudflare/cloudflared/releases/latest/download/cloudflared-linux-amd64.deb
sudo dpkg -i cloudflared-linux-amd64.deb
```

#### Setup

1. Login to Cloudflare:
   ```bash
   cloudflared tunnel login
   ```

2. Create a tunnel:
   ```bash
   cloudflared tunnel create go-radio
   ```

3. Configure the tunnel by creating `~/.cloudflared/config.yml`:
   ```yaml
   tunnel: go-radio
   credentials-file: /path/to/tunnel/credentials.json
   
   ingress:
     - hostname: your-radio.yourdomain.com
       service: http://localhost:8080
     - service: http_status:404
   ```

4. Route traffic to your tunnel:
   ```bash
   cloudflared tunnel route dns go-radio your-radio.yourdomain.com
   ```

#### Usage

1. Start your Go Radio server:
   ```bash
   make run
   ```

2. Start the tunnel:
   ```bash
   cloudflared tunnel run go-radio
   ```

### 3. localhost.run

[localhost.run](https://localhost.run/) is a simple, no-signup-required tunneling service.

#### Usage

1. Start your Go Radio server:
   ```bash
   make run
   ```

2. In a new terminal, create a tunnel using SSH:
   ```bash
   ssh -R 80:localhost:8080 localhost.run
   ```

The service will provide you with a public URL.

### 4. serveo.net

[Serveo](https://serveo.net/) is another SSH-based tunneling service.

#### Usage

1. Start your Go Radio server:
   ```bash
   make run
   ```

2. Create a tunnel:
   ```bash
   ssh -R 80:localhost:8080 serveo.net
   ```

## Security Considerations

When exposing your Go Radio instance to the internet, consider these security measures:

### 1. Authentication

Ensure you have proper authentication configured:
- Set strong admin credentials in your `.env` file:
  ```
  ADMIN_USERNAME=your_secure_username
  ADMIN_PASSWORD=your_secure_password
  ```

### 2. HTTPS

Most tunneling services provide HTTPS by default, which encrypts traffic between users and your radio.

### 3. Access Control

Consider using tunneling services that offer access control features:
- ngrok: Password protection, OAuth, IP whitelisting
- Cloudflare: Access policies, authentication

### 4. Firewall

Ensure your local firewall allows connections on port 8080 only from localhost, not from external interfaces.

## Troubleshooting

### Common Issues

1. **Port already in use**: Make sure no other service is running on port 8080
2. **Tunnel not working**: Check that your Go Radio server is running and accessible at `http://localhost:8080`
3. **Audio not playing**: Ensure your browser supports the audio format and check browser console for errors

### Testing Your Setup

1. Start Go Radio locally: `make run`
2. Test local access: Open `http://localhost:8080` in your browser
3. Start your chosen tunnel service
4. Test external access: Open the provided public URL in your browser (or ask someone else to test)

## Performance Tips

- For better performance, consider using a VPS instead of tunneling for production use
- Some tunneling services have bandwidth limits on free plans
- Audio streaming can be bandwidth-intensive, especially with multiple concurrent users

## Alternative: VPS Deployment

For production use or better performance, consider deploying Go Radio directly to a VPS using services like:
- DigitalOcean
- Linode
- AWS EC2
- Google Cloud Compute
- Vultr

This eliminates the need for tunneling and provides better performance and reliability.