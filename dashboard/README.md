# Revalyze Dashboard

Production-grade dashboard for Revalyze SaaS Pricing Intelligence platform.

## Tech Stack

- **Vite** - Build tool
- **React** - UI library (JavaScript)
- **TailwindCSS** - Styling
- **React Router** - Navigation

## Getting Started

### Prerequisites

- Node.js 18+ and npm

### Installation

```bash
cd dashboard
npm install
```

### Development

```bash
npm run dev
```

The dashboard will be available at `http://localhost:3000`

### Build for Production

```bash
npm run build
```

## Project Structure

```
dashboard/
├── src/
│   ├── components/
│   │   ├── AuthCard.jsx
│   │   └── onboarding/
│   │       └── ProgressIndicator.jsx
│   ├── pages/
│   │   ├── Login.jsx
│   │   ├── SignUp.jsx
│   │   └── onboarding/
│   │       ├── OnboardingLayout.jsx
│   │       ├── Step1Company.jsx
│   │       ├── Step2PricingContext.jsx
│   │       ├── Step3Competitors.jsx
│   │       └── Step4Stripe.jsx
│   ├── main.jsx
│   └── index.css
├── index.html
├── package.json
├── vite.config.js
├── tailwind.config.js
└── postcss.config.js
```

## Features

- Clean, enterprise-grade authentication UI
- Login and Sign Up pages
- 4-step onboarding flow with progress tracking
- Matches landing page design language
- Dark theme with blue accents
- Responsive design
- Form validation
- Smooth animations

## Design System

Matches the Revalyze landing page:
- Background: slate-900, slate-800
- Primary: blue-500, indigo-600
- Text: slate colors
- Modern, professional aesthetic

## License

© 2025 Revalyze B.V. All rights reserved.

