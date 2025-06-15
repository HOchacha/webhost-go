#!/bin/bash

echo "🛑 실행 중인 VM을 모두 종료하고 삭제합니다..."

# 현재 실행 중인 VM 목록 가져오기
running_vms=$(virsh list --name)

for vm in $running_vms; do
  echo "⏹️ VM 종료 중: $vm"
  virsh destroy "$vm"

  echo "🧹 VM 정의 삭제 중: $vm"
  virsh undefine "$vm" --remove-all-storage

  echo "✅ $vm 삭제 완료"
done

echo "🎉 모든 실행 중인 VM이 삭제되었습니다."