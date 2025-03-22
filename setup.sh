#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# Prompt the user to start the installation process
echo "🚀 Shortcut: This script will install Knative, and deploy the bookstore's frontend and backend on your cluster"

read -p "🛑 Press ENTER to continue or Ctrl+C to abort..."

# Install Knative Serving
echo ""
echo "📦 Installing Knative Serving..."
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.14.0/serving-crds.yaml
kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.14.0/serving-core.yaml
kubectl apply -f https://github.com/knative/net-kourier/releases/download/knative-v1.14.0/kourier.yaml

# Configure Kourier as the default ingress
kubectl patch configmap/config-network --namespace knative-serving --type merge --patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
echo "✅ Knative Serving installed successfully."

# Install Knative Eventing
echo ""
echo "📦 Installing Knative Eventing..."
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.14.0/eventing-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.14.0/eventing-core.yaml
echo "✅ Knative Eventing installed successfully."

# Install Knative IMC Broker
echo ""
echo "📦 Installing Knative In-Memory Channel and Broker..."
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.14.0/in-memory-channel.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/knative-v1.14.0/mt-channel-broker.yaml
echo "✅ Knative In-Memory Channel and Broker installed successfully."

# Detect whether the user has knative function "func" installed
if ! command -v func &> /dev/null
then
    echo ""
    echo "⚠️ Knative CLI 'func' not found. Please install the Knative CLI by following the instructions at https://knative.dev/docs/functions/install-func"
    exit
fi

# Wait until all pods in knative-serving and knative-eventing become ready
echo ""
echo "⏳ Waiting for Knative Serving and Eventing pods to be ready..."
kubectl wait --for=condition=ready pod --all -n knative-serving --timeout=300s
kubectl wait --for=condition=ready pod --all -n knative-eventing --timeout=300s
echo "✅ All Knative pods are ready."

# The setup is now complete
echo ""
echo "🎉 The setup is now complete."