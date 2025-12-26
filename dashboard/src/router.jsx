import { createBrowserRouter, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { CompetitorsProvider } from './context/CompetitorsContext';
import { PlansProvider } from './context/PlansContext';
import { AnalysisProvider } from './context/AnalysisV2Context';
import { SettingsProvider } from './context/SettingsContext';
import { BusinessMetricsProvider } from './context/BusinessMetricsContext';
import ProtectedRoute from './components/ProtectedRoute';
import Login from './pages/Login';
import SignUp from './pages/SignUp';
import OnboardingLayout from './pages/onboarding/OnboardingLayout';
import DashboardLayout from './layouts/DashboardLayout';
import Overview from './pages/dashboard/Overview';
import Analyses from './pages/dashboard/AnalysesV2';
import Competitors from './pages/dashboard/CompetitorsV2';
import Plans from './pages/dashboard/PlansV2'; // Using V2 as the main Plans page
import PricingSimulation from './pages/dashboard/PricingSimulation';
import Reports from './pages/dashboard/Reports';
import Settings from './pages/dashboard/Settings';
import Billing from './pages/dashboard/Billing';
import VerifyEmail from './pages/VerifyEmail';

// Wrapper component to provide all contexts for dashboard
const DashboardWithProviders = () => (
  <CompetitorsProvider>
    <PlansProvider>
      <AnalysisProvider>
        <SettingsProvider>
          <BusinessMetricsProvider>
            <DashboardLayout />
          </BusinessMetricsProvider>
        </SettingsProvider>
      </AnalysisProvider>
    </PlansProvider>
  </CompetitorsProvider>
);

// Protected dashboard wrapper
const ProtectedDashboard = () => (
  <ProtectedRoute>
    <DashboardWithProviders />
  </ProtectedRoute>
);

// Root layout with AuthProvider
const RootLayout = ({ children }) => (
  <AuthProvider>
    {children}
  </AuthProvider>
);

// Create routes with auth wrapper
export const router = createBrowserRouter([
  {
    path: '/',
    element: <RootLayout><Navigate to="/login" replace /></RootLayout>
  },
  {
    path: '/login',
    element: <RootLayout><Login /></RootLayout>
  },
  {
    path: '/signup',
    element: <RootLayout><SignUp /></RootLayout>
  },
  {
    path: '/auth/verify-email',
    element: <RootLayout><VerifyEmail /></RootLayout>
  },
  {
    path: '/onboarding',
    element: (
      <RootLayout>
        <ProtectedRoute>
          <OnboardingLayout />
        </ProtectedRoute>
      </RootLayout>
    )
  },
  {
    path: '/app',
    element: <RootLayout><ProtectedDashboard /></RootLayout>,
    children: [
      {
        index: true,
        element: <Navigate to="/app/overview" replace />
      },
      {
        path: 'overview',
        element: <Overview />
      },
      {
        path: 'analyses',
        element: <Analyses />
      },
      {
        path: 'competitors',
        element: <Competitors />
      },
      {
        path: 'plans',
        element: <Plans />
      },
      {
        path: 'simulation',
        element: <PricingSimulation />
      },
      {
        path: 'reports',
        element: <Reports />
      },
      {
        path: 'settings',
        element: <Settings />
      },
      {
        path: 'settings/billing',
        element: <Billing />
      },
      {
        path: 'billing',
        element: <Billing />
      }
    ]
  }
]);
