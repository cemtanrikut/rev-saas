const Billing = () => {
  const handleUpgradeClick = (planName) => {
    console.log(`Upgrade flow coming soon for: ${planName}`);
  };

  const plans = [
    {
      name: 'Starter',
      price: 39,
      currency: '€',
      interval: 'month',
      description: 'Perfect for early-stage startups validating pricing',
      features: [
        'Up to 3 active pricing analyses per month',
        'Up to 5 tracked competitors',
        'Email export of recommendations',
        'Community support'
      ],
      cta: 'Select Plan',
      popular: false
    },
    {
      name: 'Growth',
      price: 129,
      currency: '€',
      interval: 'month',
      description: 'For growing SaaS companies optimizing pricing',
      features: [
        'Unlimited pricing analyses',
        'Up to 20 tracked competitors',
        'Priority support',
        'Early access to new AI models',
        'Advanced reporting & exports',
        'Stripe integration (coming soon)'
      ],
      cta: 'Select Plan',
      popular: true
    },
    {
      name: 'Enterprise',
      price: 299,
      currency: '€',
      interval: 'month',
      description: 'For established companies with complex pricing',
      features: [
        'Custom limits & SLAs',
        'Dedicated success manager',
        'Security & procurement review support',
        'Multi-team collaboration',
        'API access',
        'Custom integrations'
      ],
      cta: 'Contact Us',
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
          Review your current plan and explore available upgrade options. Stripe integration is coming soon.
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
                Free Preview
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
            You are currently on a free preview environment with limited usage. Upgrades will be enabled once billing is live.
          </p>
        </div>
      </div>

      {/* Available Plans Section */}
      <div>
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-white mb-2">
            Available Plans
          </h2>
          <p className="text-slate-400">
            Choose the plan that best fits your pricing intelligence needs.
          </p>
        </div>

        {/* Pricing Cards Grid */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
          {plans.map((plan) => (
            <div
              key={plan.name}
              className={`relative bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border transition-all hover:border-slate-600 ${
                plan.popular 
                  ? 'border-blue-500/50 shadow-xl shadow-blue-500/10' 
                  : 'border-slate-800'
              }`}
            >
              {/* Popular Badge */}
              {plan.popular && (
                <div className="absolute -top-3 left-1/2 -translate-x-1/2">
                  <span className="px-4 py-1 bg-gradient-to-r from-blue-500 to-indigo-600 text-white text-xs font-semibold rounded-full shadow-lg">
                    Most Popular
                  </span>
                </div>
              )}

              {/* Plan Header */}
              <div className="mb-6">
                <h3 className="text-xl font-bold text-white mb-2">
                  {plan.name}
                </h3>
                <p className="text-sm text-slate-400 leading-relaxed">
                  {plan.description}
                </p>
              </div>

              {/* Price */}
              <div className="mb-6">
                <div className="flex items-baseline gap-1">
                  <span className="text-4xl font-bold text-white">
                    {plan.currency}{plan.price}
                  </span>
                  <span className="text-slate-400">
                    / {plan.interval}
                  </span>
                </div>
              </div>

              {/* Features List */}
              <ul className="space-y-3 mb-8">
                {plan.features.map((feature, index) => (
                  <li key={index} className="flex items-start gap-3">
                    <svg 
                      className="w-5 h-5 text-blue-400 flex-shrink-0 mt-0.5" 
                      fill="none" 
                      stroke="currentColor" 
                      viewBox="0 0 24 24"
                    >
                      <path 
                        strokeLinecap="round" 
                        strokeLinejoin="round" 
                        strokeWidth={2} 
                        d="M5 13l4 4L19 7" 
                      />
                    </svg>
                    <span className="text-slate-300 text-sm leading-relaxed">
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
        <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-800">
          <div className="flex items-start gap-4">
            <div className="w-10 h-10 bg-amber-500/10 rounded-xl flex items-center justify-center flex-shrink-0">
              <svg className="w-5 h-5 text-amber-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
            </div>
            <div>
              <h3 className="text-white font-semibold mb-2">
                Coming Soon
              </h3>
              <p className="text-slate-400 text-sm leading-relaxed">
                Self-serve upgrades will be available once we finalize Stripe integration. For now, this page is a preview of the planned plans and pricing. If you're interested in early access to a paid plan, reach out to us at{' '}
                <a href="mailto:billing@revalyze.com" className="text-blue-400 hover:text-blue-300 transition-colors">
                  billing@revalyze.com
                </a>
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* FAQ / Additional Info */}
      <div className="bg-slate-900/50 backdrop-blur-sm rounded-2xl p-8 border border-slate-800">
        <h3 className="text-lg font-semibold text-white mb-6">
          Frequently Asked Questions
        </h3>
        
        <div className="space-y-6">
          <div>
            <h4 className="text-white font-medium mb-2">
              Can I change plans later?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              Yes. Once billing is enabled, you'll be able to upgrade or downgrade at any time. Changes take effect immediately, with prorated billing adjustments.
            </p>
          </div>

          <div>
            <h4 className="text-white font-medium mb-2">
              What payment methods do you accept?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              We accept all major credit cards (Visa, Mastercard, American Express) and SEPA bank transfers for European customers. Payments are processed securely via Stripe.
            </p>
          </div>

          <div>
            <h4 className="text-white font-medium mb-2">
              Is there a free trial?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              You're currently in the free preview phase. Once billing launches, new users will have a 14-day free trial on any paid plan, no credit card required.
            </p>
          </div>

          <div>
            <h4 className="text-white font-medium mb-2">
              Do you offer annual billing?
            </h4>
            <p className="text-slate-400 text-sm leading-relaxed">
              Yes. Annual plans come with a 20% discount and will be available once billing is enabled. Contact us for enterprise volume discounts.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default Billing;


