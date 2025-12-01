import { createBrowserRouter, Navigate } from 'react-router-dom';
import Login from './pages/Login';
import SignUp from './pages/SignUp';
import OnboardingLayout from './pages/onboarding/OnboardingLayout';
import DashboardLayout from './layouts/DashboardLayout';
import Overview from './pages/dashboard/Overview';
import Analyses from './pages/dashboard/Analyses';
import Competitors from './pages/dashboard/Competitors';
import Reports from './pages/dashboard/Reports';
import Settings from './pages/dashboard/Settings';

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
    element: <DashboardLayout />,
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
        path: 'reports',
        element: <Reports />
      },
      {
        path: 'settings',
        element: <Settings />
      }
    ]
  }
]);

