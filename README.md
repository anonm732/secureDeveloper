# Go Secure Coding Practice

보안 코딩 연습을 위한 시작 프로젝트입니다.

처음부터 구조를 예쁘게 나누기보다, `cmd/server/main.go` 하나에 코드를 모아 둔 상태에서
먼저 흐름을 이해하고 직접 분리 기준을 고민할 수 있게 만드는 것이 목적입니다.

지난 과제와 수업에서 설명했던 내용, 그리고 전달된 가이드 문서를 떠올리면서
어떤 기능부터 구현하고 어떤 기준으로 나눌지 스스로 판단해 보세요.

## 프로젝트 목적

- 로그인 흐름을 먼저 이해하기
- 더미 API를 실제 동작 코드로 바꿔 보기
- 게시판과 금융 기능을 단계적으로 채워 보기
- 입력 검증, 인증 확인, 권한 검사, 응답 설계를 직접 고민해 보기
- 구현 후 어떤 코드들을 묶어 리팩터링할지 판단해 보기

## 현재 상태

- 서버 코드는 현재 `cmd/server/main.go` 하나에 들어 있습니다.
- 로그인만 SQLite 조회를 사용합니다.
- 로그인 성공 시 `authorization` 쿠키와 `Authorization` 헤더용 토큰을 함께 사용할 수 있습니다.
- 세션은 DB가 아니라 메모리 맵으로 관리합니다.
- 게시판 API는 고정 더미 응답을 반환합니다.
- 금융 API는 더미 응답을 반환합니다.
- 정적 화면은 SPA 형태로 준비되어 있습니다.

## 기본 계정

- `alice / alice1234`
- `bob / bob1234`
- `charlie / charlie1234`

## 주요 API

인증
- `POST /api/auth/register`
- `POST /api/auth/login`
- `POST /api/auth/logout`
- `POST /api/auth/withdraw`

사용자
- `GET /api/me`

게시판
- `GET /api/posts`
- `GET /api/posts/:id`
- `POST /api/posts`
- `PUT /api/posts/:id`
- `DELETE /api/posts/:id`

금융
- `POST /api/banking/deposit`
- `POST /api/banking/withdraw`
- `POST /api/banking/transfer`

## 참고 파일

- `schema.sql`
- `seed.sql`
- `query_examples.sql`

## 먼저 해볼 작업

1. 회원가입 더미 핸들러를 실제 `INSERT`로 바꿔 보기
2. 게시글 작성, 조회, 수정, 삭제를 실제 SQL 또는 원하는 저장 방식으로 바꿔 보기
3. 입금, 출금, 송금 더미 핸들러를 실제 로직으로 바꿔 보기
4. 실패 조건과 입력 검증 규칙을 정리해 보기
5. 반복되는 인증 확인 코드를 어떻게 줄일지 생각해 보기

## 작업하면서 점검할 질문

- 이 코드는 요청 처리인지, 비즈니스 규칙인지, DB 접근인지?
- 같은 인증 확인 코드가 반복되고 있지 않은가?
- 게시글 수정/삭제 권한 검사는 어디에 두는 것이 자연스러운가?
- 출금과 송금에서 어떤 실패 조건을 먼저 막아야 하는가?
- 어떤 응답은 그대로 내려도 되고, 어떤 응답은 가공이 필요한가?
- 언제부터 `handler`, `service`, `store` 같은 구조로 나누는 것이 좋은가?

## 실행 방법

프로젝트 루트에서 실행합니다.

```powershell
go run ./cmd/server
```
처음 상태로 다시 시작하고 싶으면 `app.db`를 지운 뒤 다시 실행하면 됩니다.

### 작업한 내용의 특징에 대해서 작성해주세요
[실제로 구현한 기능들]
> main.go:17~18, 로깅과 로그 로테이션을 위한 라이브러리 의존성 추가
> main.go:31, 비밀번호 해싱을 위한 Salt 필드 추가
> main.go:74~77, 게시글 작성자 정보 저장을 위한 Author, AuthorEmail 필드 추가
> main.go:119~121, 패키지 경로 분류 후 경로 갱신 및 const화
> main.go:123, log 생성 위치 추가
> main.go:132, 로깅 및 감사를 위한 initLogger() 추가
> main.go:137, 로그 로테이션을 위한 JSONLogger() 추가
> main.go:155, api/auth/register 회원가입 기능 구현(DB 데이터 삽입, 비밀번호 해싱)
> main.go:191, api/auth/login 로그인 기능 보완(해싱된 비밀번호 비교 검증)
> main.go:402, api/posts 게시글 작성 기능 구현
> main.go:438, api/posts:id 특정 게시글 조회 기능 구현
> main.go:568, 비밀번호 해싱 비교 검증을 위한 sql 구문 수정
> main.go:587, 비밀번호 유출을 방지하기 위 해싱 함수 구현
> main.go:608, 회원가입 쿼리 실행 함수 구현
> main.go:630, 게시글 작성 쿼리 실행 함수 구현
> main.go:649, 특정 id 게시글 조회 쿼리 실행 함수 구현
> main.go:670, 로깅을 위한 initLogger 함수 구현
> main.go:685, 로그 로테이션을 위한 JSONLogger 함수 구현 (충분한 로깅을 위해 MaxSize 10으로 수정)
> 체계적인 코드관리를 위해 폴더를 아래처럼 구분
>> pkg/
>>>> cmd/server/main.go
>>>> ext/db/sqlite/
>>>>>> init/
>>>>>>>> schema.sql
>>>>>>>> seed.sql
>>>>>> sample/
>>>>>>>> query_examples.sql
>>>>>> app.db
>>>> handlers
>>>> logs/
>>>> static/
>> go.mod
>> go.sum
>> README.md
> 


[구현에 있어 고려해야 할 점들]
- >> 항상 가용성을 신경쓸 것 <<
- const 적재적소에 사용하기
- 알맞은 곳에 getter/setter
- 코드 잘 읽히도록 짜기
- 구조체는 어떻게 설계?
- API는 기능별로 분리(절대 하나의 범용 API를 만들지마)
- Public 함수명은 대문자 시작, Private 함수명 소문자 시작
- 요청에 대한 검증 (파라미터, 경곅값)
- SQLi -> Prepared Statement 사용
- Path Traversal 방지 (Canonical Path)
- SSRF 발생할 건덕지 만들지 말기 (요청 받고 다른 API 호출하지 말)
- ReDoS 방지 (안전한 Regex 사)
- 로그인, 검색, 외부 API 호출 : 연속 요청 방지 로직 만들기
- 이미지 변환 : 업로드된 파일의 용량을 검사해서 일정 수준 미만으로 줄이기
- Content-Type을 application/json으로 고정 (예외 처리)
- 필요한 정보만 노출하기(DB 쿼리 결과, 에러 페이지)
- int 형에 대한 언더플로우 방지
- 길이 제한 빡세게
- 요청으로 들어오는 body 길이 limit 걸기
- 타임아웃 몇 초 설정해야할까?
- Client 요청 - Route - 바인딩 - 검증 - 서비스 동작 - 저장소(로깅) - 응답
- 허용 문자집합 설정 하거나 획일화하기(인코딩 일관성)
- 검증 로직은 검증 하나당 if 하나 (길이/상태/권한 분리)
- 로깅할 때 요청한 사용자 ID만 포함
- HTTP Method 제한
- 응답 형태 고정하기
- 미들웨어 활용하기 (인증/공통 정책, 요청 컨텍스트)
- 반대로 서비스/저장소는 권한 결정 대신하지 않기 (비즈니스 규칙/인가, 저장소 쿼리)
- CSRF 검증 어떻게 할까? (일단 후순위)
- Client-Agent가 API면 거부, 모바일이면 웹으로만 접속 가능하다고 안내 후 거부
- 로그인 성공 이후의 "인가/정보노출/자원통제"에 더 신경쓰기
- 응답 패킷에 담을 데이터는 최소 필드만. 필터링 잘 하기
- 거절(429) 정책 + 사용자 메시지 신경쓰기

- PW는 해시화해서 저장
- 회원가입 : SELECT 후 중복 사용자(ID, 탈퇴 여부) 검증 -> 중복 없으면 PW에 Slat쳐서 INSERT INTO하기
- 로그인 : SELECT로 ID, PW 조회 -> ID 비교 후에 맞으면 PW 조회 (PW는 해시화 후 비교)
- 마이페이지 : SELECT로 회원 정보 조회
- 회원 정보 갱신 : 사용자 요청 body에서 정보 파싱해서 UPDATE table SET 하기 (PW 역시 Salt쳐서 Hash)
- 로그인 시, 세션 만들 타임아웃 두기 (갱신 버튼 누르면 시간 초기화 작업하기) / 짧은 만료 : 안전하지만 재발급 빈도 Up, UX 비용 Up / 긴 만료 : 운영 편함, 탈취 피해 Up, 권한 변경 반영 Down
- 세션은 어떻게 할까? (JWT는 쓰지 말기, 지금 구현할 서비스에 알맞지 않음)
- 세션에는 최소한의 권한 정보만 두기
- 사용자 쿠키에는 HttpOnly, Secure 걸기
- 사용자에게는 세션 ID만 발급, 쿠키로 들고 있게 하기
- 로그아웃 누르면 세션 즉시 만료 + 세션 재사용 금지

- 회원 정보 테이블 (ID[prim key], 닉네임(가변 20자), <회원가입 일자>, PW Hash, Salt, <탈퇴 여부>)

db 파일처럼 외부 컴포넌트는 ext(ernal) 이라는 디렉토리로 따로 뺴서 모아둠
pkg
ㄴ dtos
ㄴ ext/
	ㄴ db/
		ㄴ mysql
ㄴ handlers/
	ㄴ static.go
ㄴ static
main.go

[계층 구조]
Hnadler : I/O, 파싱, 응답 형식
Service : 비즈니스 규칙, 인가, 상태 전이
Repository : 영속성, 쿼리, 트랜잭션

- Router / Middleware
- Handler / DTO
- Service / AuthZ
- Repository / DB

[우선 순위]
1. 요청-응답 기능
2. DB 연결과 쿼리 기능 정상 동작 확인
3. 가용성 침해 가능성 검토와 피드백 반영
4. <추가 바람>

[체크리스트]
- 입력은 무엇을 허용할 것인가?
- 상태는 어디에 둘 것인가? 즉시 회수는 필요한가?
- 출력은 누구에게 어디까지 보여줄 것인가?
- 자원은 누가 얼마나 쓰게 할 것인가?
