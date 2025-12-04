import { createBrowserRouter, Navigate } from 'react-router-dom';
import { CompetitorsProvider } from './context/CompetitorsContext';
import { PlansProvider } from './context/PlansContext';
import { AnalysisProvider } from './context/AnalysisContext';
import { SettingsProvider } from './context/SettingsContext';
import Login from './pages/Login';
import SignUp from './pages/SignUp';
import OnboardingLayout from './pages/onboarding/OnboardingLayout';
import DashboardLayout from './layouts/DashboardLayout';
import Overview from './pages/dashboard/Overview';
import Analyses from './pages/dashboard/Analyses';
import Competitors from './pages/dashboard/Competitors';
import Plans from './pages/dashboard/Plans';
import Reports from './pages/dashboard/Reports';
import Settings from './pages/dashboard/Settings';
import Billing from './pages/dashboard/Billing';

const DashboardWithProvider = () => (
  <CompetitorsProvider>
    <PlansProvider>
      <AnalysisProvider>
        <SettingsProvider>
          <DashboardLayout />
        </SettingsProvider>
      </AnalysisProvider>
    </PlansProvider>
  </CompetitorsProvider>
);

export const router = createBrowserRouter([
  {
    path: '/',
    element: <Navigate to="/login" replace />
  },
  {
    path: '/login',
    element: <Login />
  },
  {
    path: '/signup',
    element: <SignUp />
  },
  {
    path: '/onboarding',
    element: <OnboardingLayout />
  },
  {
    path: '/app',
    element: <DashboardWithProvider />,
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
        path: 'reports',
        element: <Reports />
      },
      {
        path: 'settings',
        element: <Settings />
      },
      {
        path: 'billing',
        element: <Billing />
      }
    ]
  }
]);


