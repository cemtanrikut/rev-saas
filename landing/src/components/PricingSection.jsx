import { motion } from 'framer-motion';
import { useInView } from 'framer-motion';
import { useRef } from 'react';

const PricingCard = ({ plan, price, description, features, creditNote, isPopular, delay }) => {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: "-100px" });

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5, delay }}
      className={`relative bg-slate-800/50 backdrop-blur-sm rounded-3xl p-8 flex flex-col transition-all duration-300 ${
        isPopular 
          ? 'border-2 border-blue-500 shadow-2xl shadow-blue-500/30 scale-105' 
          : 'border border-slate-700 hover:shadow-xl hover:shadow-blue-500/10 hover:border-blue-500/50'
      }`}
    >
      {isPopular && (
        <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
          <div className="bg-gradient-to-r from-blue-500 to-indigo-600 text-white px-5 py-1.5 rounded-full text-sm font-semibold shadow-lg shadow-blue-500/50">
            Most Popular
          </div>
        </div>
      )}

      <div className="mb-8">
        <h3 className="text-xl font-semibold text-white mb-3">{plan}</h3>
        <div className="flex items-baseline mb-3">
          <span className="text-5xl font-bold bg-gradient-to-r from-blue-400 to-indigo-400 bg-clip-text text-transparent">{price}</span>
          <span className="text-slate-400 ml-2">/month</span>
        </div>
        <p className="text-slate-400">{description}</p>
      </div>

      <div className="flex-grow mb-8">
        <ul className="space-y-4">
          {features.map((feature, index) => (
            <li key={index} className="flex items-start gap-3">
              <div className="flex-shrink-0 w-6 h-6 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-full flex items-center justify-center mt-0.5 shadow-sm">
                <svg className="w-3.5 h-3.5 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24" strokeWidth={3}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <div className="flex flex-col">
                <span className="text-slate-300 leading-relaxed">{feature}</span>
                {feature.includes('AI Insight Credits') && creditNote && (
                  <span className="text-sm font-medium text-blue-400 mt-1">({creditNote})</span>
                )}
              </div>
            </li>
          ))}
        </ul>
      </div>

      <button
        className={`w-full py-3.5 rounded-2xl font-semibold transition-all duration-300 ${
          isPopular
            ? 'bg-gradient-to-r from-blue-500 to-indigo-600 text-white shadow-lg shadow-blue-500/50 hover:shadow-xl hover:shadow-blue-500/60 hover:scale-105'
            : 'bg-slate-700 text-white hover:bg-slate-600 shadow-sm hover:shadow-md hover:scale-105'
        }`}
      >
        Get Started
      </button>
    </motion.div>
  );
};

const PricingSection = () => {
  const plans = [
    {
      plan: "Starter",
      price: "€69",
      description: "For small SaaS teams",
      features: [
        "Up to 3 pricing plans",
        "Up to 3 competitors",
        "AI-powered pricing analysis",
        "5 AI Insight Credits / month",
        "PDF export of pricing reports"
      ],
      creditNote: "AI analyses only",
      isPopular: false
    },
    {
      plan: "Growth",
      price: "€159",
      description: "For growing SaaS companies",
      features: [
        "Up to 5 pricing plans",
        "Up to 5 competitors",
        "AI-powered pricing analysis",
        "Pricing simulations (what-if analysis)",
        "20 AI Insight Credits / month",
        "PDF export of analyses & simulations"
      ],
      creditNote: "AI analyses + simulations",
      isPopular: true
    },
    {
      plan: "Enterprise",
      price: "€399",
      description: "For larger SaaS companies",
      features: [
        "7+ pricing plans",
        "10+ competitors",
        "Full AI-powered pricing analysis",
        "Full pricing simulations access",
        "100 AI Insight Credits / month",
        "CSV & Excel export",
        "3 team seats (multi-user access)"
      ],
      creditNote: "AI analyses + simulations",
      isPopular: false
    }
  ];

  return (
    <section id="pricing" className="py-24 md:py-32 bg-slate-900">
      <div className="max-w-7xl mx-auto px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-block px-4 py-1.5 bg-blue-500/10 border border-blue-500/30 rounded-full text-sm font-medium text-blue-400 mb-4">
            Pricing
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Simple, transparent pricing
          </h2>
          <p className="text-xl text-slate-400 max-w-2xl mx-auto">
            Choose the plan that fits your stage. Scale as you grow.
          </p>
        </motion.div>

        <div className="grid md:grid-cols-3 gap-8 max-w-6xl mx-auto mb-12">
          {plans.map((plan, index) => (
            <PricingCard
              key={index}
              plan={plan.plan}
              price={plan.price}
              description={plan.description}
              features={plan.features}
              creditNote={plan.creditNote}
              isPopular={plan.isPopular}
              delay={index * 0.1}
            />
          ))}
        </div>

        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          transition={{ duration: 0.5, delay: 0.3 }}
          viewport={{ once: true }}
          className="text-center"
        >
          <p className="text-slate-400">
            All plans include a 14-day free trial. No credit card required to start.
          </p>
        </motion.div>
      </div>
    </section>
  );
};

export default PricingSection;
