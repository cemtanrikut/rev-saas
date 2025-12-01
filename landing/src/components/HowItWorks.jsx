import { motion } from 'framer-motion';
import { useInView } from 'framer-motion';
import { useRef } from 'react';

const Step = ({ number, title, description }) => {
  const ref = useRef(null);
  const isInView = useInView(ref, { once: true, margin: "-100px" });

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 20 }}
      animate={isInView ? { opacity: 1, y: 0 } : { opacity: 0, y: 20 }}
      transition={{ duration: 0.5, delay: number * 0.1 }}
      className="group"
    >
      <div className="bg-slate-800/50 backdrop-blur-sm rounded-2xl p-6 border border-slate-700 hover:border-blue-500/50 hover:shadow-lg hover:shadow-blue-500/10 transition-all duration-300">
        <div className="flex items-start gap-5">
          {/* Number with gradient */}
          <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-blue-500 to-indigo-600 text-white rounded-2xl flex items-center justify-center font-bold text-lg shadow-lg shadow-blue-500/50 group-hover:scale-110 transition-transform">
            {number}
          </div>

          {/* Content */}
          <div className="flex-1 pt-1">
            <h3 className="text-lg font-semibold text-white mb-2 group-hover:text-blue-400 transition-colors">{title}</h3>
            <p className="text-slate-400 leading-relaxed">{description}</p>
          </div>
        </div>
      </div>
    </motion.div>
  );
};

const HowItWorks = () => {
  const steps = [
    {
      number: 1,
      title: "Connect your data sources",
      description: "Link your Stripe account and add your top competitors' pricing pages. Setup takes less than 5 minutes."
    },
    {
      number: 2,
      title: "AI analyzes and models scenarios",
      description: "Our AI processes your revenue data, customer behavior, and competitive positioning. We run hundreds of pricing simulations."
    },
    {
      number: 3,
      title: "Get actionable recommendations",
      description: "Receive specific pricing moves with projected revenue impact. No guesswork, just data-driven decisions."
    }
  ];

  return (
    <section id="how-it-works" className="py-24 md:py-32 bg-slate-800">
      <div className="max-w-4xl mx-auto px-6 lg:px-8">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          viewport={{ once: true }}
          className="text-center mb-16"
        >
          <div className="inline-block px-4 py-1.5 bg-blue-500/10 border border-blue-500/30 rounded-full text-sm font-medium text-blue-400 mb-4">
            Simple process
          </div>
          <h2 className="text-4xl md:text-5xl font-bold text-white mb-4 tracking-tight">
            Three simple steps
          </h2>
          <p className="text-xl text-slate-400">
            From connection to recommendation in minutes
          </p>
        </motion.div>

        <div className="space-y-6">
          {steps.map((step, index) => (
            <Step
              key={index}
              number={step.number}
              title={step.title}
              description={step.description}
            />
          ))}
        </div>
      </div>
    </section>
  );
};

export default HowItWorks;
