import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ProgressIndicator from '../../components/onboarding/ProgressIndicator';
import Step1Plans from './Step1Plans';
import Step2Competitors from './Step2Competitors';
import Step3BusinessMetrics from './Step3BusinessMetrics';
import Step4Review from './Step4Review';
import { postJson, businessMetricsApi } from '../../lib/apiClient';

const STEP_LABELS = [
  'Pricing Plans',
  'Competitors',
  'Business Metrics',
  'Review',
];

const OnboardingLayout = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [isLoading, setIsLoading] = useState(false);
  
  // Onboarding data state
  const [plans, setPlans] = useState([]);
  const [competitors, setCompetitors] = useState([]);
  const [businessMetrics, setBusinessMetrics] = useState({
    currency: 'USD',
    mrr: '',
    monthlyChurnRate: '',
    pricingGoal: '',
    targetArrGrowth: '',
  });

  const totalSteps = 4;

  const updateBusinessMetrics = (updated) => {
    setBusinessMetrics((prev) => ({ ...prev, ...updated }));
  };

  const validateStep = (step) => {
    switch (step) {
      case 1:
        // At least one plan is required
        if (plans.length === 0) {
          return { valid: false, message: 'Please add at least one pricing plan.' };
        }
        return { valid: true };

      case 2:
        // Competitors are optional
        return { valid: true };

      case 3:
        // MRR, Churn, and Pricing Goal are required
        if (businessMetrics.mrr === '' || businessMetrics.mrr === null) {
          return { valid: false, message: 'Please enter your MRR.' };
        }
        if (businessMetrics.monthlyChurnRate === '' || businessMetrics.monthlyChurnRate === null) {
          return { valid: false, message: 'Please enter your monthly churn rate.' };
        }
        if (!businessMetrics.pricingGoal) {
          return { valid: false, message: 'Please select your pricing goal.' };
        }
        return { valid: true };

      case 4:
        return { valid: true };

      default:
        return { valid: true };
    }
  };

  const handleNext = () => {
    const validation = validateStep(currentStep);
    if (!validation.valid) {
      alert(validation.message);
      return;
    }

    if (currentStep < totalSteps) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleBack = () => {
    if (currentStep > 1) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSkip = () => {
    // Only Step 2 (Competitors) can be skipped
    if (currentStep === 2) {
      setCurrentStep(currentStep + 1);
    }
  };

  const handleFinish = async () => {
    // Validate all required steps
    for (let step = 1; step <= 3; step++) {
      const validation = validateStep(step);
      if (!validation.valid) {
        alert(`Step ${step}: ${validation.message}`);
        setCurrentStep(step);
        return;
      }
    }

    setIsLoading(true);

    try {
      // 1. Save Plans
      for (const plan of plans) {
        await postJson('/api/plans', {
          name: plan.name,
          price: plan.price,
          currency: plan.currency,
          billing_cycle: plan.billingCycle,
        });
      }

      // 2. Save Competitors (if any)
      for (const comp of competitors) {
        // Convert plans to backend format
        const competitorPlans = comp.plans.map((p) => ({
          name: p.name,
          price: p.price,
          currency: p.currency,
          billing_cycle: p.billingCycle,
        }));

        await postJson('/api/competitors', {
          name: comp.name,
          url: comp.url || '',
          plans: competitorPlans,
        });
      }

      // 3. Save Business Metrics
      const metricsPayload = {
        currency: businessMetrics.currency,
        mrr: parseFloat(businessMetrics.mrr) || 0,
        monthly_churn_rate: parseFloat(businessMetrics.monthlyChurnRate) || 0,
        pricing_goal: businessMetrics.pricingGoal,
      };
      
      // Only include targetArrGrowth if it's set
      if (businessMetrics.targetArrGrowth !== '' && businessMetrics.targetArrGrowth !== null) {
        metricsPayload.target_arr_growth = parseFloat(businessMetrics.targetArrGrowth);
      }

      await businessMetricsApi.set(metricsPayload);

      // 4. Run first analysis
      await postJson('/api/analysis/run', {});

      // Navigate to dashboard
      navigate('/app/analyses');
    } catch (error) {
      console.error('Onboarding error:', error);
      alert('An error occurred while saving your data. Please try again.');
    } finally {
      setIsLoading(false);
    }
  };

  const renderStep = () => {
    switch (currentStep) {
      case 1:
        return <Step1Plans data={plans} onChange={setPlans} />;
      case 2:
        return <Step2Competitors data={competitors} onChange={setCompetitors} />;
      case 3:
        return <Step3BusinessMetrics data={businessMetrics} onChange={updateBusinessMetrics} />;
      case 4:
        return (
          <Step4Review
            plans={plans}
            competitors={competitors}
            businessMetrics={businessMetrics}
            onGenerate={handleFinish}
            isLoading={isLoading}
          />
        );
      default:
        return null;
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-b from-slate-900 via-slate-800 to-slate-900 flex flex-col items-center justify-center px-4 py-12">
      <div className="w-full max-w-3xl">
        {/* Logo */}
        <div className="flex justify-center mb-8 animate-fadeIn">
          <img src="/revalyze-logo.png" alt="Revalyze" className="h-14 w-auto" />
        </div>

        {/* Progress Indicator */}
        <ProgressIndicator
          currentStep={currentStep}
          totalSteps={totalSteps}
          stepLabels={STEP_LABELS}
        />

        {/* Step Content */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-3xl shadow-2xl p-10 border border-slate-700 mb-8">
          {renderStep()}
        </div>

        {/* Navigation Buttons */}
        {currentStep < totalSteps && (
          <div className="flex justify-between items-center">
            <button
              onClick={handleBack}
              disabled={currentStep === 1}
              className={`px-6 py-3 rounded-xl font-semibold transition-all ${
                currentStep === 1
                  ? 'bg-slate-700 text-slate-500 cursor-not-allowed'
                  : 'bg-slate-700 text-white hover:bg-slate-600 hover:scale-105'
              }`}
            >
              Back
            </button>

            <div className="flex items-center gap-3">
              {/* Skip button only on Step 2 */}
              {currentStep === 2 && (
                <button
                  onClick={handleSkip}
                  className="px-6 py-3 text-slate-400 hover:text-white transition-colors font-medium"
                >
                  Skip this step
                </button>
              )}

              <button
                onClick={handleNext}
                className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all"
              >
                Next
              </button>
            </div>
          </div>
        )}

        {/* Back button only on Step 4 */}
        {currentStep === totalSteps && (
          <div className="flex justify-start">
            <button
              onClick={handleBack}
              className="px-6 py-3 rounded-xl font-semibold bg-slate-700 text-white hover:bg-slate-600 hover:scale-105 transition-all"
            >
              Back
            </button>
          </div>
        )}

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-slate-400 text-sm">© {new Date().getFullYear()} Revalyze B.V.</p>
        </div>
      </div>
    </div>
  );
};

export default OnboardingLayout;
