import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../../context/AuthContext';

const Billing = () => {
  const navigate = useNavigate();
  const { isAdmin, user } = useAuth();
  
  // Get current plan name for display
  const currentPlan = user?.plan || 'free';
  const planDisplayName = {
    free: 'Free Preview',
    starter: 'Starter',
    growth: 'Growth',
    enterprise: 'Enterprise',
    admin: 'Admin'
  }[currentPlan] || 'Free Preview';

  // Redirect admin users away from billing page
  useEffect(() => {
    if (isAdmin) {
      navigate('/app/overview', { replace: true });
    }
  }, [isAdmin, navigate]);

  const handleUpgradeClick = (planName) => {
    console.log(`Upgrade flow coming soon for: ${planName}`);
  };

  // Don't render billing content for admin users
  if (isAdmin) {
    return null;
  }

  const plans = [
    {
      name: 'Starter',
      price: 69,
      currency: '€',
      interval: 'month',
      description: 'For early-stage SaaS founders',
      features: [
        '3 competitors',
        '3 pricing plans',
        '5 analyses / month',
        'AI-powered report (single scenario)',
        'PDF export'
      ],
      cta: 'Upgrade',
      popular: false
    },
    {
      name: 'Growth',
      price: 159,
      currency: '€',
      interval: 'month',
      description: 'For scaling SaaS companies',
      features: [
        '5 competitors',
        '5 pricing plans',
        '10 analyses / month',
        'AI-powered report (multi-scenario)',
        'Pricing trend insights',
        'Competitor additions',
        'CSV & Excel export'
      ],
      cta: 'Upgrade',
      popular: true
    },
    {
      name: 'Enterprise',
      price: 399,
      currency: '€',
      interval: 'month',
      description: 'For established SaaS teams',
      features: [
        '10 competitors',
        '7 pricing plans',
        '20 analyses / month',
        'Strategic AI report',
        'Priority competitor integration',
        'CSV & Excel export',
        '3 team seats'
      ],
      cta: 'Contact Sales',
      popular: false
    }
  ];

  return (
    <div className="max-w-7xl mx-auto space-y-8 pb-12">
      {/* Page Header */}
      <div>
        <h1 className="text-3xl font-bold text-white mb-2">
          Billing & Subscription
        </h1>
        <p className="text-slate-400">
          Manage your subscription and explore upgrade options.
        </p>
      </div>

      {/* Current Plan Section */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
        <div className="flex items-start justify-between mb-6">
          <div>
            <h2 className="text-xl font-semibold text-white mb-2">
              Current Plan
            </h2>
            <div className="flex items-center gap-3">
              <span className="text-2xl font-bold text-blue-400">
                {planDisplayName}
              </span>
              <span className="px-3 py-1 bg-blue-500/10 text-blue-400 rounded-full text-sm font-medium">
                Active
              </span>
            </div>
          </div>
          <div className="w-12 h-12 bg-blue-500/10 rounded-xl flex items-center justify-center">
            <svg className="w-6 h-6 text-blue-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        </div>
        
        <div className="p-4 bg-slate-800/30 rounded-xl border border-slate-700">
          <p className="text-slate-300 leading-relaxed">
            {currentPlan === 'free' 
              ? 'You are on a free preview. Upgrade to unlock more competitors, analyses, and advanced AI reports.'
              : `You are on the ${planDisplayName} plan. Upgrade anytime to unlock more features.`
            }
          </p>
        </div>
      </div>

      {/* Available Plans Section */}
      <div>
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-white mb-2">
            Choose Your Plan
          </h2>
          <p className="text-slate-400">
            Scale your pricing intelligence as your business grows.
          </p>
        </div>

        {/* Pricing Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`relative bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border transition-all hover:border-slate-600 ${
                plan.popular 
                  ? 'border-blue-500/50 shadow-xl shadow-blue-500/10' 
                  : 'border-slate-800'
              }`}
            >
              {/* Popular Badge */}
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="px-4 py-1 bg-gradient-to-r from-blue-500 to-indigo-600 text-white text-xs font-semibold rounded-full shadow-lg">
                    Recommended
                  </span>
                </div>
              )}

              {/* Plan Header */}
              <div className="mb-5">
                <h3 className="text-lg font-bold text-white mb-1">
                  {plan.name}
                </h3>
                <p className="text-sm text-slate-400">
                  {plan.description}
                </p>
              </div>

              {/* Price */}
              <div className="mb-5">
                <div className="flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-white">
                    {plan.currency}{plan.price}
                  </span>
                  <span className="text-slate-400 text-sm">
                    /{plan.interval}
                  </span>
                </div>
              </div>

              {/* Features List */}
              <ul className="space-y-2.5 mb-6">
                {plan.features.map((feature, index) => (
                  <li key={index} className="flex items-start gap-2.5">
                    <svg 
                      className="w-4 h-4 text-blue-400 flex-shrink-0 mt-0.5" 
                      fill="none" 
                      stroke="currentColor" 
                      viewBox="0 0 24 24"
                    >
                      <path 
                        strokeLinecap="round" 
                        strokeLinejoin="round" 
                        strokeWidth={2.5} 
                        d="M5 13l4 4L19 7" 
                      />
                    </svg>
                    <span className="text-slate-300 text-sm">
                      {feature}
                    </span>
                  </li>
                ))}
              </ul>

              {/* CTA Button */}
              <button
                onClick={() => handleUpgradeClick(plan.name)}
                className={`w-full py-3 rounded-xl font-semibold transition-all ${
                  plan.popular
                    ? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white hover:from-blue-600 hover:to-indigo-700 hover:scale-105 shadow-lg shadow-blue-500/20'
                    : 'bg-slate-800 text-white hover:bg-slate-700'
                }`}
              >
                {plan.cta}
              </button>
            </div>
          ))}
        </div>

        {/* Notice */}
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-5 border border-slate-800">
          <div className="flex items-start gap-4">
            <div className="w-10 h-10 bg-amber-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
              <svg className="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h3 className="text-white font-semibold mb-1">
                Stripe Integration Coming Soon
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                Self-serve upgrades will be available shortly. For early access or Enterprise inquiries, contact{' '}
                <a href="mailto:billing@revalyze.com" className="text-blue-400 hover:text-blue-300 transition-colors">
                  billing@revalyze.com
                </a>
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* FAQ */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
        <h3 className="text-lg font-semibold text-white mb-5">
          Frequently Asked Questions
        </h3>
        
        <div className="grid md:grid-cols-2 gap-6">
          <div>
            <h4 className="text-white font-medium mb-1.5">
              Can I change plans later?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              Yes. Upgrade or downgrade anytime with prorated billing.
            </p>
          </div>

          <div>
            <h4 className="text-white font-medium mb-1.5">
              What payment methods do you accept?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              All major credit cards and SEPA transfers via Stripe.
            </p>
          </div>

          <div>
            <h4 className="text-white font-medium mb-1.5">
              Is there a free trial?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              Yes. 14-day free trial on all paid plans, no card required.
            </p>
          </div>

          <div>
            <h4 className="text-white font-medium mb-1.5">
              Do you offer annual billing?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              Yes. Save 20% with annual plans (coming soon).
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Billing;
