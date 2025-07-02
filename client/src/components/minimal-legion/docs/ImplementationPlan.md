# Minimal Legion - Ability System Implementation Plan

## 🚀 Implementation Roadmap

이 문서는 GAS 스타일 어빌리티 시스템의 단계별 구현 계획을 담고 있습니다.

## 📋 Phase 1: Core Foundation (Week 1)

### 1.1 Type Definitions & Interfaces
- [ ] `AbilityTypes.ts` - 핵심 타입 정의
- [ ] `EffectTypes.ts` - 이펙트 관련 타입
- [ ] `AttributeTypes.ts` - 어트리뷰트 타입

### 1.2 Core Components
- [ ] `GameplayAttribute.ts` - 기본 어트리뷰트 시스템
- [ ] `GameplayTagSystem.ts` - 태그 관리 시스템
- [ ] `AbilitySystemComponent.ts` - 메인 ASC 클래스

### 1.3 Basic Testing
- [ ] Unit tests for attributes
- [ ] Unit tests for tag system
- [ ] Basic ASC functionality tests

**Milestone 1**: 기본 어트리뷰트와 태그 시스템이 동작하는 상태

## 📋 Phase 2: Effects System (Week 2)

### 2.1 Effect Foundation
- [ ] `GameplayEffect.ts` - 기본 이펙트 클래스
- [ ] `EffectManager.ts` - 이펙트 생명주기 관리
- [ ] `AttributeModifier.ts` - 어트리뷰트 수정 시스템

### 2.2 Effect Types
- [ ] `InstantEffect.ts` - 즉시 적용 이펙트 (데미지, 힐링)
- [ ] `DurationEffect.ts` - 지속 시간 이펙트 (버프, 디버프)
- [ ] `PeriodicEffect.ts` - 주기적 이펙트 (DoT, HoT)

### 2.3 Effect Integration
- [ ] ASC와 Effect 시스템 통합
- [ ] Effect stacking 로직
- [ ] Effect cleanup 시스템

**Milestone 2**: 다양한 이펙트가 올바르게 적용/제거되는 상태

## 📋 Phase 3: Ability System (Week 3)

### 3.1 Ability Foundation
- [ ] `GameplayAbility.ts` - 기본 어빌리티 클래스
- [ ] `AbilityContext.ts` - 어빌리티 실행 컨텍스트
- [ ] `CooldownManager.ts` - 쿨다운 관리

### 3.2 Ability Lifecycle
- [ ] Ability activation flow
- [ ] Cost checking and payment
- [ ] Targeting system
- [ ] Ability cancellation

### 3.3 Basic Abilities
- [ ] `BasicAttackAbility.ts` - 기본 공격
- [ ] `FireballAbility.ts` - 투사체 어빌리티
- [ ] `HealAbility.ts` - 자가 회복 어빌리티

**Milestone 3**: 플레이어가 기본적인 어빌리티를 사용할 수 있는 상태

## 📋 Phase 4: Integration (Week 4)

### 4.1 Player Integration
- [ ] Player 클래스에 ASC 통합
- [ ] 기존 스탯 시스템을 어트리뷰트로 마이그레이션
- [ ] 입력 시스템과 어빌리티 연결

### 4.2 Enemy Integration
- [ ] Enemy 클래스에 ASC 통합
- [ ] AI 어빌리티 사용 로직
- [ ] 기존 능력을 어빌리티로 변환

### 4.3 Visual Integration
- [ ] 어빌리티 시전 시각 효과
- [ ] 이펙트 적용 시각 피드백
- [ ] UI 요소 (쿨다운, 마나바 등)

**Milestone 4**: 기존 게임과 완전히 통합된 어빌리티 시스템

## 📋 Phase 5: Advanced Features (Week 5)

### 5.1 Complex Abilities
- [ ] `LightningChainAbility.ts` - 멀티 타겟 어빌리티
- [ ] `ShieldAbility.ts` - 보호막 어빌리티
- [ ] `SummonAbility.ts` - 소환 어빌리티

### 5.2 Advanced Effects
- [ ] Effect stacking variations
- [ ] Conditional effects
- [ ] Triggered effects (reactive abilities)

### 5.3 Performance Optimization
- [ ] Object pooling for effects
- [ ] Batch attribute updates
- [ ] Memory management

**Milestone 5**: 복잡한 어빌리티와 이펙트가 원활하게 동작하는 상태

## 📋 Phase 6: Polish & Testing (Week 6)

### 6.1 Comprehensive Testing
- [ ] Integration tests
- [ ] Performance tests
- [ ] Edge case handling

### 6.2 Documentation
- [ ] API documentation
- [ ] Usage examples
- [ ] Best practices guide

### 6.3 Final Polish
- [ ] Code cleanup and refactoring
- [ ] Error handling improvements
- [ ] Performance optimizations

**Milestone 6**: 프로덕션 준비 완료

## 🔧 Implementation Details

### 시작 순서

1. **타입 정의부터 시작**
   ```typescript
   // AbilityTypes.ts 먼저 작성
   export interface AbilityContext {
     owner: any;
     target?: any;
     scene: Phaser.Scene;
     payload?: any;
   }
   ```

2. **간단한 어트리뷰트 시스템 구현**
   ```typescript
   // 가장 기본적인 형태부터
   class GameplayAttribute {
     constructor(
       public name: string,
       public baseValue: number,
       public maxValue?: number
     ) {}
   }
   ```

3. **점진적 기능 추가**
   - 단순한 것부터 복잡한 것 순서로
   - 각 단계마다 테스트 추가
   - 기존 코드와의 호환성 유지

### 기존 시스템과의 통합 전략

#### 기존 플레이어 스탯 → 어트리뷰트 마이그레이션
```typescript
// Before (기존)
class Player {
  health = 100;
  attackPower = 25;
  moveSpeed = 100;
}

// After (어빌리티 시스템 적용)
class Player {
  abilitySystem = new AbilitySystemComponent(this);
  
  constructor() {
    this.abilitySystem.addAttribute('health', 100, 100);
    this.abilitySystem.addAttribute('attackPower', 25);
    this.abilitySystem.addAttribute('moveSpeed', 100);
  }
  
  // 기존 코드 호환성을 위한 getter
  get health() {
    return this.abilitySystem.getAttributeValue('health');
  }
}
```

#### 기존 업그레이드 시스템 → 어빌리티 시스템 연결
```typescript
// 기존 업그레이드를 이펙트로 변환
const healthUpgrade = new AttributeModifierEffect('health', 'add', 20);
player.abilitySystem.applyGameplayEffect(healthUpgrade);
```

### 테스트 전략

#### 1. 단위 테스트 (Jest)
```typescript
describe('GameplayAttribute', () => {
  test('should calculate final value correctly', () => {
    const attr = new GameplayAttribute('test', 100);
    attr.addModifier({ operation: 'add', magnitude: 20 });
    expect(attr.finalValue).toBe(120);
  });
});
```

#### 2. 통합 테스트
```typescript
describe('Ability System Integration', () => {
  test('should apply damage ability correctly', async () => {
    const player = new TestPlayer();
    const enemy = new TestEnemy();
    
    const success = await player.abilitySystem.tryActivateAbility('basic_attack', { target: enemy });
    
    expect(success).toBe(true);
    expect(enemy.health).toBeLessThan(100);
  });
});
```

#### 3. 퍼포먼스 테스트
```typescript
describe('Performance', () => {
  test('should handle 100 active effects efficiently', () => {
    const player = new TestPlayer();
    const startTime = performance.now();
    
    // 100개 이펙트 적용
    for (let i = 0; i < 100; i++) {
      player.abilitySystem.applyGameplayEffect(testEffect);
    }
    
    // 업데이트 실행
    player.abilitySystem.update(16); // 60fps
    
    const endTime = performance.now();
    expect(endTime - startTime).toBeLessThan(5); // 5ms 이내
  });
});
```

## 🔍 Risk Mitigation

### 잠재적 문제점과 대응책

1. **성능 문제**
   - **위험**: 많은 이펙트로 인한 프레임 드롭
   - **대응**: Object pooling, 배치 업데이트, 우선순위 시스템

2. **복잡성 증가**
   - **위험**: 시스템이 너무 복잡해져서 유지보수 어려움
   - **대응**: 단계별 구현, 충분한 문서화, 단순한 API 설계

3. **기존 시스템과의 충돌**
   - **위험**: 기존 코드와 호환성 문제
   - **대응**: 점진적 마이그레이션, 호환성 레이어 제공

4. **메모리 누수**
   - **위험**: 이펙트나 어빌리티 인스턴스가 제대로 정리되지 않음
   - **대응**: 명확한 생명주기 관리, 자동 정리 시스템

## 📈 Success Metrics

### 구현 성공 지표

1. **기능적 지표**
   - ✅ 모든 기존 게임 기능이 어빌리티 시스템으로 동작
   - ✅ 새로운 어빌리티 추가가 10분 이내 가능
   - ✅ 복잡한 어빌리티 조합이 올바르게 동작

2. **성능 지표**
   - ✅ 60fps 유지 (100개 이상의 활성 이펙트 상황에서)
   - ✅ 메모리 사용량 20% 이내 증가
   - ✅ 로딩 시간 증가 없음

3. **코드 품질 지표**
   - ✅ 90% 이상 테스트 커버리지
   - ✅ TypeScript 에러 0개
   - ✅ ESLint 경고 0개

4. **사용성 지표**
   - ✅ 새로운 개발자가 2시간 내에 새 어빌리티 작성 가능
   - ✅ 문서만 보고도 시스템 이해 가능
   - ✅ 디버깅 도구 완비

---

*이 계획서는 구현 과정에서 발견되는 이슈에 따라 조정될 수 있습니다.*