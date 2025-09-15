local status="300"
if [[ "${status}" -gt 400 ]]; then
  echo "greater 400"
else
  echo "smaller 400"
fi
