import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import ProgressIndicator from '../../components/onboarding/ProgressIndicator';
import Step1Company from './Step1Company';
import Step2PricingContext from './Step2PricingContext';
import Step3Competitors from './Step3Competitors';
import Step4Stripe from './Step4Stripe';

const OnboardingLayout = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [onboardingData, setOnboardingData] = useState({
    company: {
      companyName: '',
      website: '',
      country: '',
      teamSize: '',
      productCategory: ''
    },
    pricingContext: {
      planCount: '',
      mainPricingConcern: '',
      pricingNotes: ''
    },
    competitors: {
      competitor1: '',
      competitor2: '',
      competitor3: ''
    },
    stripe: {
      willConnectStripe: false
    }
  });

  const totalSteps = 4;

  const updateCompany = (updated) => {
    setOnboardingData(prev => ({
      ...prev,
      company: { ...prev.company, ...updated }
    }));
  };

  const updatePricingContext = (updated) => {
    setOnboardingData(prev => ({
      ...prev,
      pricingContext: { ...prev.pricingContext, ...updated }
    }));
  };

  const updateCompetitors = (updated) => {
    setOnboardingData(prev => ({
      ...prev,
      competitors: { ...prev.competitors, ...updated }
    }));
  };

  const updateStripe = (updated) => {
    setOnboardingData(prev => ({
      ...prev,
      stripe: { ...prev.stripe, ...updated }
    }));
  };

  const validateStep = (step) => {
    switch (step) {
      case 1:
        const { companyName, website, country, teamSize, productCategory } = onboardingData.company;
        if (!companyName || !website || !country || !teamSize || !productCategory) {
          return false;
        }
        if (!website.startsWith('http://') && !website.startsWith('https://')) {
          alert('Website must start with http:// or https://');
          return false;
        }
        return true;
      
      case 2:
        const { planCount, mainPricingConcern } = onboardingData.pricingContext;
        if (!planCount || !mainPricingConcern) {
          return false;
        }
        return true;
      
      case 3:
        // All competitors are optional
        return true;
      
      case 4:
        return true;
      
      default:
        return true;
    }
  };

  const handleNext = () => {
    if (!validateStep(currentStep)) {
      alert('Please fill in all required fields.');
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

  const handleFinish = () => {
    // Validate all steps
    for (let step = 1; step <= 3; step++) {
      if (!validateStep(step)) {
        alert(`Please complete all required fields in Step ${step}.`);
        setCurrentStep(step);
        return;
      }
    }

    console.log('Onboarding data:', onboardingData);
    // TODO: Send to backend
    
    // Redirect to dashboard
    navigate('/app/overview');
  };

  const renderStep = () => {
    switch (currentStep) {
      case 1:
        return (
          <Step1Company
            data={onboardingData.company}
            onChange={updateCompany}
          />
        );
      case 2:
        return (
          <Step2PricingContext
            data={onboardingData.pricingContext}
            onChange={updatePricingContext}
          />
        );
      case 3:
        return (
          <Step3Competitors
            data={onboardingData.competitors}
            onChange={updateCompetitors}
          />
        );
      case 4:
        return (
          <Step4Stripe
            data={onboardingData.stripe}
            onChange={updateStripe}
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
          <img 
            src="/revalyze-logo.png" 
            alt="Revalyze" 
            className="h-14 w-auto"
          />
        </div>

        {/* Progress Indicator */}
        <ProgressIndicator currentStep={currentStep} totalSteps={totalSteps} />

        {/* Step Content */}
        <div className="bg-slate-800/50 backdrop-blur-sm rounded-3xl shadow-2xl p-10 border border-slate-700 mb-8">
          {renderStep()}
        </div>

        {/* Navigation Buttons */}
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

          {currentStep < totalSteps ? (
            <button
              onClick={handleNext}
              className="px-6 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all"
            >
              Next
            </button>
          ) : (
            <button
              onClick={handleFinish}
              className="px-8 py-3 bg-gradient-to-r from-blue-500 to-indigo-600 text-white rounded-xl font-semibold hover:from-blue-600 hover:to-indigo-700 hover:scale-105 hover:shadow-xl hover:shadow-blue-500/30 transition-all"
            >
              Finish onboarding
            </button>
          )}
        </div>

        {/* Footer */}
        <div className="mt-8 text-center">
          <p className="text-slate-400 text-sm">
            © {new Date().getFullYear()} Revalyze B.V.
          </p>
        </div>
      </div>
    </div>
  );
};

export default OnboardingLayout;

